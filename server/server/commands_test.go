package server

import (
	"context"
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

func TestSetCommandToVar(t *testing.T) {
	err := setCommandToVar("MYVAR=$(echo hello)")
	assert.NoError(t, err)
	value, exists := os.LookupEnv("MYVAR")
	assert.True(t, exists)
	assert.Equal(t, "hello", value)
}

func TestSetVar(t *testing.T) {
	err := setVar("MYVAR=hello")
	assert.NoError(t, err)
	value, exists := os.LookupEnv("MYVAR")
	assert.True(t, exists)
	assert.Equal(t, "hello", value)
}

func TestRunCommand(t *testing.T) {
	t.Run("good", func(t *testing.T) {
		err := runCommand("echo hello")
		assert.NoError(t, err)
	})
	t.Run("good", func(t *testing.T) {
		err := runCommand("adsadsa")
		assert.ErrorContains(t, err, "error executing command")
	})
}

func TestStopCurrentCommand(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	commandManager := newCommandManager(logger)
	ctx, cancel := context.WithCancel(context.Background())
	commandManager.SetCancelCommand(cancel)
	go func() { _ = commandManager.ExecuteCommand(ctx, "sleep 5") }()
	time.Sleep(1 * time.Second) // Ensure the command is running
	assert.True(t, commandManager.IsRunning())

	err := commandManager.StopCurrentCommand()
	assert.NoError(t, err)
	assert.False(t, commandManager.IsRunning())
}

func TestTermClear(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	commandManager := newCommandManager(logger)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		assert.NoError(t, err)
		commandManager.SetWebsocketConnection(ws)
	}))
	defer server.Close()

	ws, _, err := websocket.DefaultDialer.Dial(strings.Replace(server.URL, "http", "ws", 1), nil)
	assert.NoError(t, err)
	defer ws.Close()
	commandManager.SetWebsocketConnection(ws)

	err = commandManager.TermClear()
	assert.NoError(t, err)
}

func TestTermMessage(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	commandManager := newCommandManager(logger)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		assert.NoError(t, err)
		commandManager.SetWebsocketConnection(ws)
	}))
	defer server.Close()

	ws, _, err := websocket.DefaultDialer.Dial(strings.Replace(server.URL, "http", "ws", 1), nil)
	assert.NoError(t, err)
	defer ws.Close()
	commandManager.SetWebsocketConnection(ws)

	err = commandManager.TermMessage([]byte("test message"))
	assert.NoError(t, err)
}

func TestExecuteCommand(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	commandManager := newCommandManager(logger)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		assert.NoError(t, err)
		commandManager.SetWebsocketConnection(ws)
	}))

	ws, _, err := websocket.DefaultDialer.Dial(strings.Replace(server.URL, "http", "ws", 1), nil)
	assert.NoError(t, err)
	defer ws.Close()
	commandManager.SetWebsocketConnection(ws)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = commandManager.ExecuteCommand(ctx, "echo hello")
	assert.NoError(t, err)
}

func TestStartCommand(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	commandManager := newCommandManager(logger)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		assert.NoError(t, err)
		commandManager.SetWebsocketConnection(ws)
	}))

	ws, _, err := websocket.DefaultDialer.Dial(strings.Replace(server.URL, "http", "ws", 1), nil)
	assert.NoError(t, err)
	defer ws.Close()
	commandManager.SetWebsocketConnection(ws)

	err = commandManager.StartCommand("echo hello")
	assert.NoError(t, err)
}

func TestWebSocketConnection(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	commandManager := newCommandManager(logger)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		assert.NoError(t, err)
		commandManager.SetWebsocketConnection(ws)
		err = commandManager.StartCommand("echo hello")
		require.NoError(t, err)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	assert.NoError(t, err)
	defer ws.Close()

	_, message, err := ws.ReadMessage()
	assert.NoError(t, err)
	assert.Contains(t, string(message), "\x1b[2J\x1b[H") // clearing terminal first

	_, message, err = ws.ReadMessage()
	assert.NoError(t, err)
	assert.Contains(t, string(message), "hello")
}
