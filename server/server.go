package server

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func Start(commandsFile string) error {
	m, err := NewDemoManager(commandsFile)
	if err != nil {
		return err
	}

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	r := mux.NewRouter()
	r.HandleFunc(indexEndpoint, m.indexHandler)
	r.HandleFunc(initEndpoint, m.initHandler)
	r.HandleFunc(incPageEndpoint, m.incPageHandler)
	r.HandleFunc(decPageEndpoint, m.decPageHandler)
	r.HandleFunc(setPageEndpoint, m.setPageHandler)
	r.HandleFunc(executeEndpoint, m.executeCommandHandler).
		Methods(http.MethodPost)
	r.HandleFunc(executeEndpoint, m.executeStatusHandler).
		Methods(http.MethodGet)
	r.HandleFunc(stopEndpoint, m.stopCommandHandler)

	http.Handle(indexEndpoint, r)
	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)

	return nil
}
