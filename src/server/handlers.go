package server

import (
	"fmt"
	"net/http"
	"strconv"
)

func (s *server) HandlerIndex(w http.ResponseWriter, r *http.Request) {
	err := indexHTML(contentDiv(0, s.GetSlideCount(), s.GetSlide(0), false)).Render(w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.logger.Error("could not execute index handler", "error", err.Error())
		return
	}
}

// this allows for using the select dropdown to change pages with no extra javascript https://htmx.org/examples/value-select/
func (s *server) HandlerSlideByQuery(w http.ResponseWriter, r *http.Request) {
	if slideIdx := r.URL.Query().Get("idx"); slideIdx != "" {
		http.Redirect(w, r, fmt.Sprintf("/slides/%v", slideIdx), http.StatusMovedPermanently)
		return
	}

	w.WriteHeader(http.StatusBadRequest)
}

func (s *server) HandlerSlideByIndex(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.logger.Error("could not parse path parameter 'id' in slide handler", "error", err.Error())
		return
	}

	err = s.commandManager.termClear()
	if err != nil {
		s.logger.Warn("could not clear terminal in slide handler", "error", err.Error())
	}

	err = s.commandManager.stopCurrentCommand()
	if err != nil {
		s.logger.Warn("could not stop current command in slide handler", "command", s.GetSlide(id).content, "error", err.Error())
	}

	err = contentDiv(id, s.GetSlideCount(), s.GetSlide(id), false).Render(w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.logger.Error("could not execute slide handler", "error", err.Error())
		return
	}
}

func (s *server) HandlerInit(w http.ResponseWriter, r *http.Request) {
	s.logger.Info("attempting to upgrade HTTP to websocket")

	if s.commandManager.ws != nil {
		s.logger.Info("closing existing websocket")
		if err := s.commandManager.ws.Close(); err != nil {
			s.logger.Warn("error closing existing websocket", "error", err.Error())
		}
	}

	var err error
	s.commandManager.ws, err = upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.logger.Error("couldn't upgrade to websocket", "error", err.Error())
		return
	}

	s.logger.Info("upgraded HTTP connection to websocket")
}

func (s *server) HandlerCommandStart(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.logger.Error("could not parse path parameter 'id' in command start handler", "error", err.Error())
		return
	}

	err = runningButton(id, true).Render(w)
	if err != nil {
		s.logger.Warn("could not render running button in command start handler", "running", true, "error", err.Error())
	}

	err = s.commandManager.startCommand(s.GetSlide(id).content)
	if err != nil {
		s.logger.Error("could not start command in command start handler", "command", s.GetSlide(id).content, "error", err.Error())
	}
}

func (s *server) HandlerCommandStatus(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.logger.Error("could not parse path parameter 'id' in command status handler", "error", err.Error())
		return
	}

	if !s.commandManager.running.Load() {
		err = runningButton(id, false).Render(w)
		if err != nil {
			s.logger.Warn("could not render running button in command status handler", "running", false, "error", err.Error())
		}
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *server) HandlerCommandStop(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.logger.Error("could not parse path parameter 'id' in command stop handler", "error", err.Error())
		return
	}

	err = s.commandManager.stopCurrentCommand()
	if err != nil {
		s.logger.Warn("could not stop current command in stop command handler", "command", s.GetSlide(id).content, "error", err.Error())
	}

	err = runningButton(id, false).Render(w)
	if err != nil {
		s.logger.Warn("could not render running button in stop command handler", "running", false, "error", err.Error())
	}
}
