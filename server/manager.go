package server

import (
	"context"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync/atomic"

	"github.com/gorilla/websocket"
)

type DemoManager struct {
	ws            *websocket.Conn
	commands      []string
	cmd           atomic.Pointer[exec.Cmd]
	cmdNumber     atomic.Int32
	cmdContext    context.Context
	cancelCommand func()
}

func NewDemoManager(commandsFile string) (*DemoManager, error) {
	contents, err := os.ReadFile("commands.txt")
	if err != nil {
		return nil, err
	}

	var commands []string
	split := strings.Split(regexp.MustCompile(`\\\s*\n`).ReplaceAllString(string(contents), ""), "\n")
	for i := range split {
		if strings.TrimSpace(split[i]) != "" {
			commands = append(commands, split[i])
		}
	}

	return &DemoManager{
		commands: commands,
	}, nil
}
