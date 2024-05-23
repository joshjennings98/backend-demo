package server

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync/atomic"

	"github.com/gorilla/websocket"
)

const (
	terminalBufferSize = 1024
)

var (
	varPattern            = regexp.MustCompile(`^\w+=\w+$`)
	varCommandPattern     = regexp.MustCompile(`^(\w+)=(\$\((.+)\))$`)
	varCommandWithComment = regexp.MustCompile(`^(.*)#(.*)$`)

	upgrader = websocket.Upgrader{
		ReadBufferSize:  terminalBufferSize,
		WriteBufferSize: terminalBufferSize,
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow connections from any origin
		},
	}
)

type commandManager struct {
	cancelCommand context.CancelFunc
	running       atomic.Bool
	ws            *websocket.Conn
	logger        *slog.Logger
}

func newCommandManager(logger *slog.Logger) ICommandManager {
	return &commandManager{
		logger: logger,
	}
}

func setCommandToVar(ctx context.Context, cmd string) error {
	if matches := varCommandPattern.FindStringSubmatch(cmd); len(matches) == 4 {
		variable, command := matches[1], matches[3]

		out, err := exec.CommandContext(ctx, "sh", "-c", command).Output()
		if err != nil {
			return fmt.Errorf("error executing command '%v': %v", command, err.Error())
		}

		err = os.Setenv(variable, strings.TrimSpace(string(out)))
		if err != nil {
			return fmt.Errorf("error setting env var '%v': %v", cmd, err.Error())
		}
	}

	return nil
}

func setVar(cmd string) error {
	parts := strings.SplitN(cmd, "=", 2)

	err := os.Setenv(parts[0], parts[1])
	if err != nil {
		return fmt.Errorf("error setting env var '%v': %v", cmd, err.Error())
	}

	return nil
}

func runCommand(ctx context.Context, cmd string) error {
	err := exec.CommandContext(ctx, "sh", "-c", cmd).Run()
	if err != nil {
		return fmt.Errorf("error executing command '%v': %v", cmd, err.Error())
	}

	return nil
}

func (c *commandManager) IsRunning() bool {
	return c.running.Load()
}

func (c *commandManager) SetRunning(b bool) {
	c.running.Store(b)
}

func (c *commandManager) GetWebsocketConnection() *websocket.Conn {
	return c.ws
}

func (c *commandManager) SetWebsocketConnection(ws *websocket.Conn) {
	c.ws = ws
}

func (c *commandManager) SetCancelCommand(cancel context.CancelFunc) {
	c.cancelCommand = cancel
}

func (c *commandManager) StopCurrentCommand() (err error) {
	if c.cancelCommand != nil {
		c.cancelCommand()
		c.cancelCommand = nil
	}
	c.running.Store(false)
	return
}

func (c *commandManager) TermClear() (err error) {
	if c.ws == nil {
		return errors.New("websocket connection not yet ready")
	}
	if err := c.ws.WriteMessage(websocket.TextMessage, []byte("\033[2J\033[H")); err != nil {
		return fmt.Errorf("error sending clear to websocket: %v", err.Error())
	}
	return nil
}

func (c *commandManager) TermMessage(message []byte) error {
	if c.ws == nil {
		return errors.New("websocket connection not yet ready")
	}
	messageStr := strings.ReplaceAll(strings.TrimPrefix(strings.TrimPrefix(string(message), "sh: line 1: "), "sh: 1: "), "\n", "\n\r")
	if err := c.ws.WriteMessage(websocket.TextMessage, []byte(messageStr)); err != nil {
		return fmt.Errorf("error sending clear to websocket: %v", err.Error())
	}
	return nil
}

func (c *commandManager) ExecuteCommand(ctx context.Context, command string) (err error) {
	cmd := exec.CommandContext(ctx, "sh", "-c", command)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	c.running.Store(true)
	defer func() {
		c.running.Store(false)
	}()

	reader := bufio.NewReader(io.MultiReader(stdoutPipe, stderrPipe))
	buffer := make([]byte, terminalBufferSize)

	for {
		select {
		case <-ctx.Done():
			return nil // command context cancelled
		default:
			n, err := reader.Read(buffer)
			if err != nil {
				if err != io.EOF && !errors.Is(err, fs.ErrClosed) {
					return fmt.Errorf("error reading from pipe: %v", err.Error())
				}
				return nil // command complete
			}

			if err := c.TermMessage(buffer[:n]); err != nil {
				return err
			}
		}
	}
}

func (c *commandManager) StartCommand(command string) (err error) {
	if err = c.StopCurrentCommand(); err != nil {
		return
	}

	if c.ws == nil {
		return
	}

	if err := c.TermClear(); err != nil {
		c.logger.Warn("error sending clear to websocket:", "error", err.Error())
	}

	ctx, cancel := context.WithCancel(context.Background())
	c.cancelCommand = cancel

	go func(ctx context.Context, command string) {
		c.logger.Info("starting command", "command", command)
		if err := c.ExecuteCommand(ctx, command); err != nil {
			c.logger.Error("error executing command", "command", command, "error", err.Error())
			return
		}
	}(ctx, command)

	return
}
