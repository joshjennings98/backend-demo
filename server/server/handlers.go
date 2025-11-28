package server

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
)

var upgrader = newUpgrader(false)

func (s *server) HandlerIndex(w http.ResponseWriter, r *http.Request) {
	slide, err := s.GetSlide(0)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.logger.Error("could not get first slide", "error", err.Error())
		return
	}

	err = indexHTML(contentDiv(0, s.GetSlideCount(), slide, false)).Render(w)
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

	slide, err := s.GetSlide(id)
	if err != nil {
		if errors.Is(err, ErrSlideIndexOutOfBounds) {
			w.WriteHeader(http.StatusNotFound)
			s.logger.Warn("slide index out of bounds", "id", id)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			s.logger.Error("could not get slide", "id", id, "error", err.Error())
		}
		return
	}

	// stop running command and clear terminal when changing slides
	_ = s.commandManager.Stop()
	_ = s.commandManager.Clear()

	err = contentDiv(id, s.GetSlideCount(), slide, false).Render(w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.logger.Error("could not execute slide handler", "error", err.Error())
		return
	}
}

func (s *server) monitorWebsocket(ws *websocket.Conn) {
	defer func() {
		s.logger.Info("websocket connection closed")
		_ = s.commandManager.CloseWebsocketConnection()
	}()

	for {
		// ReadMessage blocks until a message is received or connection closes so we use this to detect disconnects
		if _, _, err := ws.ReadMessage(); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				s.logger.Warn("websocket closed unexpectedly", "error", err.Error())
			}
			return
		}
	}
}

func (s *server) HandlerWebSocket(w http.ResponseWriter, r *http.Request) {
	s.logger.Info("websocket connection requested")
	err := s.commandManager.CloseWebsocketConnection()
	if err != nil {
		s.logger.Warn("error closing existing websocket", "error", err.Error())
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Error("couldn't upgrade to websocket", "error", err.Error())
		return
	}

	s.commandManager.SetWebsocketConnection(ws)
	s.logger.Info("websocket connection established")

	go s.monitorWebsocket(ws)
}

func (s *server) HandlerCommandStart(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.logger.Error("could not parse path parameter 'id' in command start handler", "error", err.Error())
		return
	}

	slide, err := s.GetSlide(id)
	if err != nil {
		if errors.Is(err, ErrSlideIndexOutOfBounds) {
			w.WriteHeader(http.StatusNotFound)
			s.logger.Warn("slide index out of bounds in command start", "id", id)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			s.logger.Error("could not get slide in command start", "id", id, "error", err.Error())
		}
		return
	}

	err = runningButton(id, true).Render(w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.logger.Error("could not render running button", "error", err.Error())
		return
	}

	err = s.commandManager.Run(slide.ExecuteContent)
	if err != nil {
		s.logger.Error("could not start commands", "error", err.Error())
	}
}

func (s *server) HandlerCommandStatus(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.logger.Error("could not parse path parameter 'id' in command status handler", "error", err.Error())
		return
	}

	_, err = s.GetSlide(id)
	if err != nil {
		if errors.Is(err, ErrSlideIndexOutOfBounds) {
			w.WriteHeader(http.StatusNotFound)
			s.logger.Warn("slide index out of bounds in command status", "id", id)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			s.logger.Error("could not get slide in command status", "id", id, "error", err.Error())
		}
		return
	}

	if !s.commandManager.IsRunning() {
		err = runningButton(id, false).Render(w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.logger.Error("could not render running button in command status handler", "running", false, "error", err.Error())
			return
		}

		return
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

	_, err = s.GetSlide(id)
	if err != nil {
		if errors.Is(err, ErrSlideIndexOutOfBounds) {
			w.WriteHeader(http.StatusNotFound)
			s.logger.Warn("slide index out of bounds in command stop", "id", id)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			s.logger.Error("could not get slide in command stop", "id", id, "error", err.Error())
		}
		return
	}

	_ = s.commandManager.Stop()

	err = runningButton(id, false).Render(w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.logger.Error("could not render running button in command stop", "error", err.Error())
		return
	}
}
