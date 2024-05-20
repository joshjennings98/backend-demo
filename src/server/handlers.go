package server

import (
	"log"
	"net/http"
	"strconv"
)

func (s *server) indexHandler(w http.ResponseWriter, r *http.Request) {
	_ = indexHTML(contentDiv(0, s.totalSlides(), s.getSlide(0), false)).Render(w)
}

func (s *server) showSlideHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_ = s.commandManager.termClear()

	_ = contentDiv(id, s.totalSlides(), s.getSlide(id), false).Render(w)
}

func (s *server) initHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("attempting to upgrade HTTP to websocket")

	if s.commandManager.ws != nil {
		log.Println("closing existing websocket")
		if err := s.commandManager.ws.Close(); err != nil {
			log.Println("error closing existing websocket:", err.Error())
		}
	}

	var err error
	s.commandManager.ws, err = upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("error upgrading to websocket:", err.Error())
		return
	}

	log.Println("upgraded HTTP connection to websocket")
}

func (s *server) startCommandHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_ = runningButton(id, true).Render(w)

	_ = s.commandManager.startCommand(s.getSlide(id).content)

}

func (s *server) statusCommandHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if running := s.commandManager.running.Load(); !running {
		_ = runningButton(id, false).Render(w)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *server) stopCommandHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_ = s.commandManager.stopCurrentCommand()

	_ = runningButton(id, false).Render(w)
}
