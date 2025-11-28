package server

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/joshjennings98/backend-demo/server/v2/mocks"
	"github.com/joshjennings98/backend-demo/server/v2/types"
)

func setupServer(t *testing.T) (*server, *mocks.MockICommandManager) {
	ctrl := gomock.NewController(t)
	mockCommandManager := mocks.NewMockICommandManager(ctrl)

	return &server{
		slides: []types.Slide{
			{ID: 0, Content: "hello", SlideType: types.SlideTypePlain},
			{ID: 1, Content: "echo world", ExecuteContent: []string{"echo world"}, SlideType: types.SlideTypeCommand},
		},
		commandManager: mockCommandManager,
		logger:         slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}, mockCommandManager
}

func TestHandlerIndex(t *testing.T) {
	s, _ := setupServer(t)
	handler := http.HandlerFunc(s.HandlerIndex)

	req, err := http.NewRequest("GET", "/", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestHandlerSlideByQuery(t *testing.T) {
	s, _ := setupServer(t)
	handler := http.HandlerFunc(s.HandlerSlideByQuery)

	t.Run("Valid query parameter", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/slide?idx=1", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusMovedPermanently, rr.Code)
		assert.Equal(t, "/slides/1", rr.Header().Get("Location"))
	})

	t.Run("Missing query parameter", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/slide", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestHandlerSlideByIndex(t *testing.T) {
	s, cmdManager := setupServer(t)

	cmdManager.
		EXPECT().
		Stop().
		Return(nil)

	cmdManager.
		EXPECT().
		Clear().
		Return(nil)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /slides/{id}", s.HandlerSlideByIndex)

	t.Run("Valid path parameter", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/slides/1", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Invalid path parameter", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/slides/invalid", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestHandlerWebSocket_BadRequest_NoUpgrade(t *testing.T) {
	s, cmdManager := setupServer(t)

	handler := http.HandlerFunc(s.HandlerWebSocket)

	cmdManager.
		EXPECT().
		CloseWebsocketConnection().
		Return(nil)

	req, err := http.NewRequest("GET", "/ws", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandlerCommandStart(t *testing.T) {
	s, cmdManager := setupServer(t)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /commands/{id}/start", s.HandlerCommandStart)

	cmdManager.
		EXPECT().
		Run(gomock.Any()).
		Return(nil)

	t.Run("Valid path parameter", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/commands/1/start", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Invalid path parameter", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/commands/invalid/start", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestHandlerCommandStatus(t *testing.T) {
	t.Run("Running", func(t *testing.T) {
		s, cmdManager := setupServer(t)

		cmdManager.
			EXPECT().
			IsRunning().
			Return(true)

		mux := http.NewServeMux()
		mux.HandleFunc("GET /commands/{id}/status", s.HandlerCommandStatus)

		t.Run("Valid path parameter", func(t *testing.T) {
			req, err := http.NewRequest("GET", "/commands/1/status", nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusNoContent, rr.Code)
		})

		t.Run("Invalid path parameter", func(t *testing.T) {
			req, err := http.NewRequest("GET", "/commands/invalid/status", nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusBadRequest, rr.Code)
		})
	})

	t.Run("Not Running", func(t *testing.T) {
		s, cmdManager := setupServer(t)

		cmdManager.
			EXPECT().
			IsRunning().
			Return(false)

		mux := http.NewServeMux()
		mux.HandleFunc("GET /commands/{id}/status", s.HandlerCommandStatus)

		req, err := http.NewRequest("GET", "/commands/1/status", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestHandlerCommandStop(t *testing.T) {
	s, cmdManager := setupServer(t)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /commands/{id}/stop", s.HandlerCommandStop)

	cmdManager.
		EXPECT().
		Stop().
		Return(nil)

	t.Run("Valid path parameter", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/commands/1/stop", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Invalid path parameter", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/commands/invalid/stop", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}
