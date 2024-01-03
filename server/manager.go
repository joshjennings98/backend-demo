package server

import (
	"context"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
)

type DemoManager struct {
	commands       []string
	cmd            *exec.Cmd
	cmdMutex       sync.Mutex
	cmdNumber      int
	cmdNumberMutex sync.Mutex
	cmdContext     context.Context
	cancelCommand  func()
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
		commands:       commands,
		cmd:            nil,
		cmdMutex:       sync.Mutex{},
		cmdNumber:      0,
		cmdNumberMutex: sync.Mutex{},
		cmdContext:     nil,
		cancelCommand:  nil,
	}, nil
}
