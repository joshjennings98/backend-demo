package server

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
)

const (
	terminalBufferSize = 1024
)

func newUpgrader(isTest bool) websocket.Upgrader {
	return websocket.Upgrader{
		ReadBufferSize:  terminalBufferSize,
		WriteBufferSize: terminalBufferSize,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			if isTest && origin == "" {
				return true
			}
			return strings.HasPrefix(origin, "http://localhost:") ||
				strings.HasPrefix(origin, "http://127.0.0.1:")
		},
	}
}

// wsWriter is a context-aware io.Writer that writes to a WebSocket connection
type wsWriter struct {
	ctx  context.Context
	conn *websocket.Conn
	mu   *sync.Mutex
}

func newWSWriter(ctx context.Context, ws *websocket.Conn, mu *sync.Mutex) *wsWriter {
	return &wsWriter{
		ctx:  ctx,
		conn: ws,
		mu:   mu,
	}
}

func (w *wsWriter) writeMessage(msg []byte) (err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.conn.WriteMessage(websocket.TextMessage, msg)
}

func (w *wsWriter) Write(p []byte) (n int, err error) {
	select {
	case <-w.ctx.Done():
		err = w.ctx.Err()
		return
	default:
	}

	if w.conn == nil {
		err = io.ErrClosedPipe
		return
	}

	msg := bytes.ReplaceAll(p, []byte{'\n'}, []byte{'\n', '\r'})
	err = w.writeMessage(msg)
	if err != nil {
		err = fmt.Errorf("could not write message '%v': %w", string(msg), err)
		return
	}

	n = len(p)
	return
}

type commandManager struct {
	cancel  context.CancelFunc
	running atomic.Bool
	ws      *websocket.Conn
	wsMu    sync.Mutex
	logger  *slog.Logger
}

func newCommandManager(logger *slog.Logger) ICommandManager {
	return &commandManager{
		logger: logger,
	}
}

func (c *commandManager) IsRunning() bool {
	return c.running.Load()
}

func (c *commandManager) SetWebsocketConnection(ws *websocket.Conn) {
	c.wsMu.Lock()
	defer c.wsMu.Unlock()
	c.ws = ws
}

func (c *commandManager) CloseWebsocketConnection() (err error) {
	c.wsMu.Lock()
	defer c.wsMu.Unlock()

	if c.ws != nil {
		err = c.ws.Close()
		c.ws = nil
		return
	}

	return
}

func (c *commandManager) GetWebsocketConnection() *websocket.Conn {
	c.wsMu.Lock()
	defer c.wsMu.Unlock()
	return c.ws
}

func (c *commandManager) Stop() (err error) {
	if c.cancel != nil {
		c.logger.Info("cancelling command")
		c.cancel()
		c.cancel = nil
	}

	c.running.Store(false)
	return
}

func (c *commandManager) Clear() (err error) {
	c.wsMu.Lock()
	defer c.wsMu.Unlock()

	if c.ws == nil {
		return
	}

	return c.ws.WriteMessage(websocket.TextMessage, []byte("\033[2J\033[H"))
}

func (c *commandManager) run(ctx context.Context, ws *websocket.Conn, script string) {
	c.running.Store(true)
	defer func() {
		c.running.Store(false)
	}()

	c.logger.Info("executing commands", "script", script)

	cmd := exec.CommandContext(ctx, "sh", "-c", script)

	writer := newWSWriter(ctx, ws, &c.wsMu)
	cmd.Stdout = writer
	cmd.Stderr = writer

	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			c.logger.Info("command stopped", "reason", ctx.Err())
		} else {
			c.logger.Error("command failed", "error", err)
		}
	} else {
		c.logger.Info("command completed")
	}
}

func (c *commandManager) Run(commands []string) (err error) {
	_ = c.Stop()

	ws := c.GetWebsocketConnection()
	if ws == nil {
		return
	}

	if err := c.Clear(); err != nil {
		c.logger.Warn("failed to clear terminal", "error", err)
	}

	var ctx context.Context
	ctx, c.cancel = context.WithCancel(context.Background())

	script := strings.Join(commands, "\n")
	go c.run(ctx, ws, script)

	return
}
