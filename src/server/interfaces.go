package server

import (
	"context"
	"net/http"
)

type IServer interface {
	initialise(ctx context.Context) error
	parsePreCommands(contents []string) (i int, err error)
	getPreCommands() []string
	parseSlides(contents []string, startIdx int) (err error)
	addSlide(content string, t slideType)
	getSlide(idx int) *slide
	totalSlides() int
	indexHandler(w http.ResponseWriter, r *http.Request)
	initHandler(w http.ResponseWriter, r *http.Request)
	showSlideHandler(w http.ResponseWriter, r *http.Request)
	startCommandHandler(w http.ResponseWriter, r *http.Request)
	statusCommandHandler(w http.ResponseWriter, r *http.Request)
	stopCommandHandler(w http.ResponseWriter, r *http.Request)
}
