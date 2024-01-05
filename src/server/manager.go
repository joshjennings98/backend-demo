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

// TODO: handle code blocks with highlight.js
var (
	commandRegex    = regexp.MustCompile(`^\$\s*`)
	textLinkRegex   = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	imageLinkRegex  = regexp.MustCompile(`^!\[([^\]]*)\]\(([^)]+)\)$`) // only allow image only slides
	inlineCodeRegex = regexp.MustCompile("`([^`]*)`")
	hrRegex         = regexp.MustCompile(`\-\-\-+`)
)

type DemoManager struct {
	ws            *websocket.Conn
	preCommands   []string
	commands      []string
	cmd           atomic.Pointer[exec.Cmd]
	cmdNumber     atomic.Int32
	cmdContext    context.Context
	cancelCommand func()
}

func NewDemoManager(commandsFile string) (*DemoManager, error) {
	contents, err := os.ReadFile(commandsFile)
	if err != nil {
		return nil, err
	}

	isPreCommands := true
	var commands, preCommands []string
	split := strings.Split(regexp.MustCompile(`\\\s*\n`).ReplaceAllString(string(contents), ""), "\n")
	for i := range split {
		if s := split[i]; strings.TrimSpace(s) != "" {
			if hrRegex.MatchString(s) {
				isPreCommands = false
				continue
			}

			if isPreCommands {
				preCommands = append(preCommands, split[i])
			} else {
				commands = append(commands, split[i])
			}
		}
	}

	return &DemoManager{
		preCommands: preCommands,
		commands:    commands,
	}, nil
}

func (m *DemoManager) termClear() error {
	if err := m.ws.WriteMessage(websocket.TextMessage, []byte("\033[2J\033[H")); err != nil {
		return fmt.Errorf("Error sending clear to websocket: %v", err.Error())
	}
	return nil
}

func (m *DemoManager) termMessage(message []byte) error {
	messageStr := strings.ReplaceAll(string(message), "\n", "\n\r")
	if err := m.ws.WriteMessage(websocket.TextMessage, []byte(messageStr)); err != nil {
		return fmt.Errorf("Error sending clear to websocket: %v", err.Error())
	}
	return nil
}

func (m *DemoManager) isCmdRunning() bool {
	return m.cmd.Load() != nil
}

func (m *DemoManager) setCommand(i int32) {
	m.cmdNumber.Store(i)
}

func (m *DemoManager) incCommand() {
	m.cmdNumber.Add(1)
}

func (m *DemoManager) decCommand() {
	m.cmdNumber.Add(-1)
}

func (m *DemoManager) getCommand() int32 {
	// https://stackoverflow.com/questions/43018206/modulo-of-negative-integers-in-go
	numCommands := int32(len(m.commands))
	return (m.cmdNumber.Load()%numCommands + numCommands) % numCommands
}

func (m *DemoManager) logInfo(_ http.ResponseWriter, format string, v ...any) {
	log.Printf(format, v...)
}

func (m *DemoManager) logError(w http.ResponseWriter, format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	log.Print(msg)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(msg))
}

func (m *DemoManager) cleanedCommand() string {
	contentString := m.commands[m.getCommand()]
	if m.isCommand() {
		return commandRegex.ReplaceAllString(contentString, "")
	}

	var output string
	output = imageLinkRegex.ReplaceAllStringFunc(contentString, func(match string) string {
		submatches := imageLinkRegex.FindStringSubmatch(match)
		return fmt.Sprintf(`<img src="%v" alt="%v">`, submatches[2], submatches[1])
	})
	output = textLinkRegex.ReplaceAllStringFunc(output, func(match string) string {
		submatches := textLinkRegex.FindStringSubmatch(match)
		return fmt.Sprintf(`<a href="%v" target="_blank" rel="noopener noreferrer">%v</a>`, submatches[2], submatches[1])
	})
	output = inlineCodeRegex.ReplaceAllStringFunc(output, func(match string) string {
		submatches := inlineCodeRegex.FindStringSubmatch(match)
		return fmt.Sprintf(`<code>%v</code>`, submatches[1])
	})

	return output
}

func (m *DemoManager) waitForCommandExit() {
	cmd := m.cmd.Load()
	if cmd == nil {
		return
	}

	var exitErr *exec.ExitError
	if err := cmd.Wait(); errors.As(err, &exitErr) {
		return // command exited
	}

	select {
	case <-m.cmdContext.Done():
		return // already cancelled
	default:
		_ = m.stopCurrentCommand()
		return
	}
}

func (m *DemoManager) startCmd() error {
	cmd := m.cmd.Load()
	if cmd == nil {
		return nil
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error starting command '%v': %v", m.cleanedCommand(), err.Error())
	}

	return nil
}

func (m *DemoManager) isCommand() bool {
	return strings.HasPrefix(m.commands[m.getCommand()], "$")
}

func (m *DemoManager) executeCommand(stdoutPipe io.Reader) error {
	if err := m.startCmd(); err != nil {
		return err
	}

	go m.waitForCommandExit()

	reader := bufio.NewReader(stdoutPipe)
	buffer := make([]byte, terminalBufferSize)

	for {
		select {
		case <-m.cmdContext.Done():
			return nil // command context cancelled
		default:
			n, err := reader.Read(buffer)
			if err != nil {
				if err != io.EOF && !errors.Is(err, fs.ErrClosed) {
					return fmt.Errorf("error reading from StdoutPipe: %v\n", err.Error())
				}
				return nil // command complete
			}

			if err := m.termMessage(buffer[:n]); err != nil {
				return err
			}
		}
	}
}

func (m *DemoManager) stopCurrentCommand() error {
	if m.cancelCommand != nil {
		m.cancelCommand()
		m.cancelCommand = nil
	}

	cmd := m.cmd.Load()
	if cmd == nil {
		return nil
	}

	if cmd.Process != nil && (cmd.ProcessState == nil || !cmd.ProcessState.Exited()) {
		if err := cmd.Process.Signal(os.Interrupt); err != nil {
			return err
		}
	}
	m.cmd.Store(nil)

	m.logInfo(nil, "stopped command '%v'", m.cleanedCommand())
	return nil
}
