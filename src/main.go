package main

import (
	"flag"
	"log"

	"github.com/joshjennings98/backend-demo/src/server"
)

func main() {
	var commands string
	flag.StringVar(&commands, "commands", "", "Commands file")
	flag.StringVar(&commands, "c", "", "Commands file")
	flag.Parse()

	if commands == "" {
		log.Fatal("commands file must be provided via -c/-commands")
	}

	err := server.Start(commands)
	if err != nil {
		log.Fatal(err)
	}
}
