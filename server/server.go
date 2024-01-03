package server

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/gorilla/mux"
)

func Start(commandsFile string) error {
	contents, err := os.ReadFile("commands.txt")
	if err != nil {
		return err
	}

	split := strings.Split(regexp.MustCompile(`\\\s*\n`).ReplaceAllString(string(contents), ""), "\n")
	for i := range split {
		if strings.TrimSpace(split[i]) != "" {
			commands = append(commands, split[i])
		}
	}

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	r := mux.NewRouter()
	r.HandleFunc(indexEndpoint, indexHandler)
	r.HandleFunc(initEndpoint, initHandler)
	r.HandleFunc(incPageEndpoint, incPageHandler)
	r.HandleFunc(decPageEndpoint, decPageHandler)
	r.HandleFunc(setPageEndpoint, setPageHandler)
	r.HandleFunc(executeEndpoint, executeCommandHandler).
		Methods(http.MethodPost)
	r.HandleFunc(executeEndpoint, executeStatusHandler).
		Methods(http.MethodGet)
	r.HandleFunc(stopEndpoint, stopCommandHandler)

	http.Handle(indexEndpoint, r)
	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)

	return nil
}
