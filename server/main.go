package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"os"

	"github.com/joshjennings98/backend-demo/server/server"
)

func main() {
	var commandsFile string
	flag.StringVar(&commandsFile, "commands", "", "Commands file")
	flag.StringVar(&commandsFile, "c", "", "Commands file")
	flag.Parse()

	if commandsFile == "" {
		log.Fatal("commands file must be provided via -c/-commands")
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	s, err := server.NewServer(logger, commandsFile)
	if err != nil {
		log.Fatal(err)
	}

	err = s.Start(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}
