package server

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync/atomic"

	"github.com/gorilla/websocket"
)

const (
	terminalBufferSize = 1024
)

var (
	varPattern        = regexp.MustCompile(`^\w+=\w+$`)
	varCommandPattern = regexp.MustCompile(`^(\w+)=(\$\((.+)\))$`)

	upgrader = websocket.Upgrader{
		ReadBufferSize:  terminalBufferSize,
		WriteBufferSize: terminalBufferSize,
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow connections from any origin
		},
	}
)

type commandManager struct {
	currentCommand string
	cancelCommand  context.CancelFunc
	running        atomic.Bool
	ws             *websocket.Conn
}

func setCommandToVar(cmd string) error {
	if matches := varCommandPattern.FindStringSubmatch(cmd); len(matches) == 4 {
		variable, command := matches[1], matches[3]

		out, err := exec.Command("sh", "-c", command).Output()
		if err != nil {
			return fmt.Errorf("error executing command '%v': %v", command, err.Error())
		}

		err = os.Setenv(variable, strings.TrimSpace(string(out)))
		if err != nil {
			return fmt.Errorf("error setting env var '%v': %v", cmd, err.Error())
		}
	}

	return nil
}

func setVar(cmd string) error {
	parts := strings.SplitN(cmd, "=", 2)

	err := os.Setenv(parts[0], parts[1])
	if err != nil {
		return fmt.Errorf("error setting env var '%v': %v", cmd, err.Error())
	}

	return nil
}

func runCommand(cmd string) error {
	err := exec.Command("sh", "-c", cmd).Run()
	if err != nil {
		return fmt.Errorf("error executing command '%v': %v", cmd, err.Error())
	}

	return nil
}

func (c *commandManager) stopCurrentCommand() (err error) {
	if c.cancelCommand != nil {
		c.cancelCommand()
		c.cancelCommand = nil
	}
	c.running.Store(false)
	return
}

func (c *commandManager) termClear() (err error) {
	if err := c.ws.WriteMessage(websocket.TextMessage, []byte("\033[2J\033[H")); err != nil {
		return fmt.Errorf("Error sending clear to websocket: %v", err.Error())
	}
	return nil
}

func (c *commandManager) termMessage(message []byte) error {
	messageStr := strings.ReplaceAll(strings.TrimPrefix(string(message), "sh: line 1: "), "\n", "\n\r")
	if err := c.ws.WriteMessage(websocket.TextMessage, []byte(messageStr)); err != nil {
		return fmt.Errorf("Error sending clear to websocket: %v", err.Error())
	}
	return nil
}

func (c *commandManager) executeCommand(ctx context.Context, command string) (err error) {
	cmd := exec.CommandContext(ctx, "sh", "-c", command)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("error creating stdout pipe for command %v: %v\n", cmd.String(), err.Error())
		return
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		log.Printf("error creating stderr pipe for command %v: %v\n", cmd.String(), err.Error())
		return
	}

	if err := cmd.Start(); err != nil {
		return err
	}
	c.running.Store(true)
	defer func() {
		c.running.Store(false)
	}()

	reader := bufio.NewReader(io.MultiReader(stdoutPipe, stderrPipe))
	buffer := make([]byte, terminalBufferSize)

	for {
		select {
		case <-ctx.Done():
			return nil // command context cancelled
		default:
			n, err := reader.Read(buffer)
			if err != nil {
				if err != io.EOF && !errors.Is(err, fs.ErrClosed) {
					return fmt.Errorf("error reading from pipe: %v\n", err.Error())
				}
				return nil // command complete
			}

			if err := c.termMessage(buffer[:n]); err != nil {
				return err
			}
		}
	}
}

func (c *commandManager) startCommand(command string) (err error) {
	if err = c.stopCurrentCommand(); err != nil {
		log.Printf("error stopping command %v: %v\n", c.currentCommand, err.Error())
		return
	}

	if c.ws == nil {
		log.Println("websocket connection not yet ready")
		return
	}

	if err := c.termClear(); err != nil {
		log.Println("error sending clear to websocket:", err.Error())
	}

	ctx, cancel := context.WithCancel(context.Background())
	c.cancelCommand = cancel

	go func(ctx context.Context, command string) {
		log.Println("starting command:", command)
		if err := c.executeCommand(ctx, command); err != nil {
			log.Printf("error executing command %v: %v\n", command, err.Error())
		}
	}(ctx, command)

	return
}
