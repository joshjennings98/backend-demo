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
	r.HandleFunc(pageEndpoint, m.incPageHandler).
		Methods(http.MethodPost)
	r.HandleFunc(pageEndpoint, m.decPageHandler).
		Methods(http.MethodDelete) // At least `PUT /page` is imdempotent :)
	r.HandleFunc(pageEndpoint, m.setPageHandler).
		Methods(http.MethodPut)
	r.HandleFunc(executeEndpoint, m.executeCommandHandler).
		Methods(http.MethodPost)
	r.HandleFunc(executeEndpoint, m.executeStatusHandler).
		Methods(http.MethodGet)
	r.HandleFunc(executeEndpoint, m.stopCommandHandler).
		Methods(http.MethodDelete)

	http.Handle(indexEndpoint, r)
	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)

	return nil
}
