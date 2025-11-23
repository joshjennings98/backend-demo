package server

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupWebSocket(t *testing.T, cm ICommandManager) (ws *websocket.Conn, cleanup func()) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := newUpgrader(true)
		ws, err := upgrader.Upgrade(w, r, nil)
		assert.NoError(t, err)
		cm.SetWebsocketConnection(ws)
	}))

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	ws, resp, err := websocket.DefaultDialer.Dial(wsURL, nil) //nolint:bodyclose // the body is closed in cleanup
	require.NoError(t, err)

	cleanup = func() {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
		_ = ws.Close()
		server.Close()
	}

	time.Sleep(10 * time.Millisecond)
	return
}

func TestCommandManager_Stop(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cm := newCommandManager(logger)

	_, cleanup := setupWebSocket(t, cm)
	defer cleanup()

	// start a long-running command
	err := cm.Run([]string{"sleep 10"})
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)
	require.True(t, cm.IsRunning())

	// stop should cancel the command
	err = cm.Stop()
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond)
	assert.False(t, cm.IsRunning())
}

func TestCommandManager_Clear(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cm := newCommandManager(logger)

	_, cleanup := setupWebSocket(t, cm)
	defer cleanup()

	err := cm.Clear()
	assert.NoError(t, err)
}

func TestCommandManager_Run(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cm := newCommandManager(logger)

	ws, cleanup := setupWebSocket(t, cm)
	defer cleanup()

	err := cm.Run([]string{"echo hello"})
	require.NoError(t, err)

	// read the clear message
	_, msg, err := ws.ReadMessage()
	require.NoError(t, err)
	assert.Contains(t, string(msg), "\033[2J")

	// read the output
	_, msg, err = ws.ReadMessage()
	require.NoError(t, err)
	assert.Contains(t, string(msg), "hello")
}

func TestCommandManager_RunMultipleCommands(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cm := newCommandManager(logger)

	ws, cleanup := setupWebSocket(t, cm)
	defer cleanup()

	// run multiple commands meaning env var should persist
	err := cm.Run([]string{"export FOO=bar", "echo $FOO"})
	require.NoError(t, err)

	_, _, err = ws.ReadMessage() // clear
	require.NoError(t, err)

	// output should contain "bar" since commands run in same shell
	_, msg, err := ws.ReadMessage()
	require.NoError(t, err)
	assert.Contains(t, string(msg), "bar")
}

func TestCommandManager_WebSocketConnection(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cm := newCommandManager(logger)

	assert.Nil(t, cm.GetWebsocketConnection())

	_, cleanup := setupWebSocket(t, cm)
	defer cleanup()

	assert.NotNil(t, cm.GetWebsocketConnection())

	err := cm.CloseWebsocketConnection()
	require.NoError(t, err)
	assert.Nil(t, cm.GetWebsocketConnection())
}
