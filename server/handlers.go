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
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
)

// TODO: handle code blocks with highlight.js
var (
	commandRegex    = regexp.MustCompile(`^\$\s*`)
	textLinkRegex   = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	imageLinkRegex  = regexp.MustCompile(`^!\[([^\]]*)\]\(([^)]+)\)$`) // only allow image only slides
	inlineCodeRegex = regexp.MustCompile("`([^`]*)`")
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  terminalBufferSize,
		WriteBufferSize: terminalBufferSize,
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow connections from any origin
		},
	}
)

const (
	indexEndpoint   = "/"
	initEndpoint    = "/init"
	pageEndpoint    = "/page"
	executeEndpoint = "/execute"

	terminalBufferSize = 1024
)

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

func (m *DemoManager) Log(_ http.ResponseWriter, format string, v ...any) {
	log.Printf(format, v...)
}

func (m *DemoManager) LogError(w http.ResponseWriter, format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	log.Print(msg)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(msg))
}

func (m *DemoManager) indexHandler(w http.ResponseWriter, r *http.Request) {
	if err := m.stopCurrentCommand(); err != nil {
		m.LogError(w, "error stopping command %v: %v", m.cleanedCommand(), err.Error())
		return
	}

	m.setCommand(0)
	m.indexHTML(m.commands, m.cmdNumber.Load(), m.isCommand(), m.isCmdRunning()).Render(w)
}

func (m *DemoManager) initHandler(w http.ResponseWriter, r *http.Request) {
	m.Log(w, "attempting to upgrade HTTP to websocket")

	if m.ws != nil {
		m.Log(w, "Closing existing websocket")
		if err := m.ws.Close(); err != nil {
			m.LogError(w, "error closing existing websocket: %v", err.Error())
		}
	}

	var err error
	m.ws, err = upgrader.Upgrade(w, r, nil)
	if err != nil {
		m.LogError(w, "error upgrading to websocket: %v", err.Error())
		return
	}

	m.Log(w, "Upgraded HTTP connection to websocket")
}

func (m *DemoManager) incPageHandler(w http.ResponseWriter, r *http.Request) {
	prevCommand := m.getCommand()
	m.incCommand()

	if err := m.stopCurrentCommand(); err != nil {
		m.LogError(w, "error stopping command %v: %v", m.cleanedCommand(), err.Error())
		return
	}

	if currCommand := m.getCommand(); prevCommand != currCommand {
		m.contentDiv(m.commands, currCommand, m.isCommand(), false).Render(w)
	} else {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err := m.termClear(); err != nil {
		m.LogError(w, "Error sending clear to websocket: %v", err.Error())
		return
	}
}

func (m *DemoManager) decPageHandler(w http.ResponseWriter, r *http.Request) {
	prevCommand := m.getCommand()
	m.decCommand()

	if err := m.stopCurrentCommand(); err != nil {
		m.LogError(w, "error stopping command %v: %v", m.cleanedCommand(), err.Error())
		return
	}

	if currCommand := m.getCommand(); prevCommand != currCommand {
		m.contentDiv(m.commands, currCommand, m.isCommand(), false).Render(w)
	} else {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err := m.termClear(); err != nil {
		m.LogError(w, "error sending clear to websocket: %v", err.Error())
		return
	}
}

func (m *DemoManager) setPageHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		m.LogError(w, "failed to parse form: %v", err.Error())
		return
	}

	slideIndex, err := strconv.Atoi(r.FormValue("slideIndex"))
	if err != nil {
		m.LogError(w, "failed to parse slide index: %v", err.Error())
		return
	}

	prevCommand := m.getCommand()
	m.setCommand(int32(slideIndex))
	isCmdRunning := m.isCmdRunning()

	if currCommand := m.getCommand(); prevCommand != currCommand {
		if err := m.stopCurrentCommand(); err != nil {
			m.LogError(w, "error stopping command %v: %v", m.cleanedCommand(), err.Error())
			return
		}

		m.contentDiv(m.commands, currCommand, m.isCommand(), isCmdRunning).Render(w)
	}

	if err := m.termClear(); err != nil {
		m.LogError(w, "error sending clear to websocket: %v", err.Error())
	}
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

func (m *DemoManager) executeCommandHandler(w http.ResponseWriter, r *http.Request) {
	if !m.isCommand() {
		return
	}

	if err := m.stopCurrentCommand(); err != nil {
		m.LogError(w, "error stopping command %v: %v", m.cleanedCommand(), err.Error())
		return
	}

	log.Printf("Starting command '%v'\n", m.cleanedCommand())

	if err := m.termClear(); err != nil {
		m.LogError(w, "error sending clear to websocket: %v", err.Error())
	}

	var cmd *exec.Cmd
	m.cmdContext, m.cancelCommand = context.WithCancel(context.Background())
	cmd = exec.CommandContext(m.cmdContext, "sh", "-c", m.cleanedCommand())
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		m.LogError(w, "error creating stdout pipe for command %v: %v", m.cleanedCommand(), err.Error())
		return
	}
	cmd.Stderr = cmd.Stdout
	m.cmd.Store(cmd)

	runningButton(true).Render(w)

	go func() {
		if err := m.executeCommand(stdoutPipe); err != nil {
			m.LogError(w, "error executing command %v: %v", m.cleanedCommand(), err.Error())
		}
		return
	}()
}

func (m *DemoManager) executeStatusHandler(w http.ResponseWriter, r *http.Request) {
	if running := m.isCmdRunning(); !running {
		runningButton(running).Render(w)
	} else {
		w.WriteHeader(http.StatusNoContent)
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

	log.Printf("Stopped command '%v'\n", m.cleanedCommand())
	return nil
}

func (m *DemoManager) stopCommandHandler(w http.ResponseWriter, r *http.Request) {
	if err := m.stopCurrentCommand(); err != nil {
		m.LogError(w, "error stopping command %v: %v", m.cleanedCommand(), err.Error())
		return
	}

	runningButton(false).Render(w)
}
