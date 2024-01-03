package main

import (
	"log"

	"github.com/joshjennings98/backend-demo/server"
)

func main() {
	err := server.Start("commands.txt")
	if err != nil {
		log.Fatal(err)
	}
}
