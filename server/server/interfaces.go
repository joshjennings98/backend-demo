package server

import (
	"context"
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/joshjennings98/backend-demo/v2/server/types"
)

type IPresentation interface {
	Initialise(ctx context.Context) error
	SplitContent(commandsFile string) (preCommands, slideContent []string, err error)
	ParsePreCommands(contents []string) (err error)
	GetPreCommands() []string
	ParseSlides(contents []string) (err error)
	ParseSlide(content string)
	GetSlide(idx int) types.Slide
	GetSlideCount() int
}

type IPresentationServer interface {
	IPresentation
	Start(ctx context.Context) error
	HandlerIndex(w http.ResponseWriter, r *http.Request)
	HandlerInit(w http.ResponseWriter, r *http.Request)
	HandlerSlideByIndex(w http.ResponseWriter, r *http.Request)
	HandlerSlideByQuery(w http.ResponseWriter, r *http.Request)
	HandlerCommandStart(w http.ResponseWriter, r *http.Request)
	HandlerCommandStatus(w http.ResponseWriter, r *http.Request)
	HandlerCommandStop(w http.ResponseWriter, r *http.Request)
}

//go:generate mockgen -destination=../mocks/mock_$GOPACKAGE.go -package=mocks github.com/joshjennings98/backend-demo/v2/server/$GOPACKAGE ICommandManager

type ICommandManager interface {
	IsRunning() bool
	SetRunning(b bool)
	SetCancelCommand(cancel context.CancelFunc)
	GetWebsocketConnection() *websocket.Conn
	SetWebsocketConnection(ws *websocket.Conn)
	StopCurrentCommand() error
	TermClear() error
	TermMessage(message []byte) error
	ExecuteCommand(ctx context.Context, command string) error
	StartCommand(command string) error
}
