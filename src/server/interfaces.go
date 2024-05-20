package server

import (
	"context"
	"net/http"
)

type IPresentation interface {
	Initialise(ctx context.Context) error
	ParsePreCommands(contents []string) (i int, err error)
	GetPreCommands() []string
	ParseSlides(contents []string, startIdx int) (err error)
	ParseSlide(content string, t slideType)
	GetSlide(idx int) slide
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
