package server

import (
	"context"
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/joshjennings98/backend-demo/server/v2/types"
)

type IPresentation interface {
	SplitContent(commandsFile string) (slideContent []string, err error)
	ParseSlides(contents []string)
	ParseSlide(content string)
	GetSlide(idx int) (types.Slide, error)
	GetSlideCount() int
}

type IPresentationServer interface {
	IPresentation
	Start(ctx context.Context) error
	HandlerIndex(w http.ResponseWriter, r *http.Request)
	HandlerWebSocket(w http.ResponseWriter, r *http.Request)
	HandlerSlideByIndex(w http.ResponseWriter, r *http.Request)
	HandlerSlideByQuery(w http.ResponseWriter, r *http.Request)
	HandlerCommandStart(w http.ResponseWriter, r *http.Request)
	HandlerCommandStatus(w http.ResponseWriter, r *http.Request)
	HandlerCommandStop(w http.ResponseWriter, r *http.Request)
}

//go:generate go tool mockgen -destination=../mocks/mock_$GOPACKAGE.go -package=mocks github.com/joshjennings98/backend-demo/server/v2/$GOPACKAGE ICommandManager

type ICommandManager interface {
	GetWebsocketConnection() *websocket.Conn
	SetWebsocketConnection(ws *websocket.Conn)
	CloseWebsocketConnection() error
	Run(commands []string) error
	Stop() error
	Clear() error
	IsRunning() bool
}
