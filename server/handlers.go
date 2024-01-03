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
	"sync"

	"github.com/gorilla/websocket"
)

// TODO: handle code blocks with highlight.js
// TODO: in the text blocks replace IMAGE ONLY with image, inline code with <code>, and links with <a>
var commandRegex = regexp.MustCompile(`^\$\s*`)
var imageRegex = regexp.MustCompile(`!\[(.*?)\]\((.*?)\)`)
var linkRegex = regexp.MustCompile(`\[(.*?)\]\((.*?)\)`)
var inlineCodeRegex = regexp.MustCompile("`(.*?)`")

var (
	upgradeOnce sync.Once
	upgrader    = websocket.Upgrader{
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
	incPageEndpoint = "/inc-page"
	decPageEndpoint = "/dec-page"
	setPageEndpoint = "/set-page"
	executeEndpoint = "/execute"
	stopEndpoint    = "/stop"

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
	m.cmdMutex.Lock()
	defer m.cmdMutex.Unlock()
	return m.cmd != nil
}

func (m *DemoManager) setCommand(i int, increment bool) int {
	m.cmdNumberMutex.Lock()
	defer m.cmdNumberMutex.Unlock()
	prevCommand := m.cmdNumber
	baseVal := 0
	if increment {
		baseVal = prevCommand
	}
	m.cmdNumber = max(0, min(len(m.commands)-1, baseVal+i))
	return prevCommand
}

func (m *DemoManager) incCommand() int {
	return m.setCommand(1, true)
}

func (m *DemoManager) decCommand() int {
	return m.setCommand(-1, true)
}

func (m *DemoManager) indexHandler(w http.ResponseWriter, r *http.Request) {
	m.stopCurrentCommand()
	m.setCommand(0, false)
	indexHTML(m.commands, m.cmdNumber, m.isCmdRunning()).Render(w)
}

func (m *DemoManager) initHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Attempting to upgrade HTTP to WebSocket")

	if m.ws != nil {
		log.Println("Closing existing websocket")
		err := m.ws.Close()
		if err != nil {
			log.Println("Error closing existing websocket:", err)
		}
	}

	var err error
	m.ws, err = upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading to websocket:", err)
		http.Error(w, "Could not open WebSocket connection", http.StatusInternalServerError)
		return
	}

	log.Println("Upgraded HTTP connection to websocket")
}

func (m *DemoManager) incPageHandler(w http.ResponseWriter, r *http.Request) {
	prevCommand := m.incCommand()

	m.stopCurrentCommand()

	if prevCommand != m.cmdNumber {
		contentDiv(m.commands, m.cmdNumber, false).Render(w)
	} else {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err := m.termClear(); err != nil {
		log.Println("Error sending clear to websocket:", err)
	}
}

func (m *DemoManager) decPageHandler(w http.ResponseWriter, r *http.Request) {
	prevCommand := m.decCommand()

	m.stopCurrentCommand()

	if prevCommand != m.cmdNumber {
		contentDiv(m.commands, m.cmdNumber, false).Render(w)
	} else {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err := m.termClear(); err != nil {
		log.Println("Error sending clear to websocket:", err)
	}
}

func (m *DemoManager) setPageHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	slideIndex := r.FormValue("slideIndex")
	slideIndexVal, err := strconv.Atoi(slideIndex)
	if err != nil {
		http.Error(w, "Invalid slide index", http.StatusBadRequest)
		return
	}

	prevCommand := m.setCommand(slideIndexVal, false)
	isCmdRunning := m.isCmdRunning()

	if m.cmdNumber != prevCommand {
		m.stopCurrentCommand()
		contentDiv(m.commands, m.cmdNumber, isCmdRunning).Render(w)
	}

	if err := m.termClear(); err != nil {
		log.Println("Error sending clear to websocket:", err)
	}
}

func (m *DemoManager) cleanedCommand() string {
	return commandRegex.ReplaceAllString(m.commands[m.cmdNumber], "")
}

func (m *DemoManager) waitForCommandExit() {
	if m.cmd == nil {
		return
	}

	var exitErr *exec.ExitError
	if err := m.cmd.Wait(); errors.As(err, &exitErr) {
		return // command exited
	}

	select {
	case <-m.cmdContext.Done():
		return // already cancelled
	default:
		m.stopCurrentCommand()
		return
	}
}

func (m *DemoManager) startCmd() error {
	m.cmdMutex.Lock()
	defer m.cmdMutex.Unlock()

	if m.cmd == nil {
		return nil
	}

	if err := m.cmd.Start(); err != nil {
		return fmt.Errorf("Error starting command '%v': %v", m.cleanedCommand(), err.Error())
	}

	return nil
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
					return fmt.Errorf("Error reading from StdoutPipe: %v\n", err.Error())
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
	m.stopCurrentCommand()
	log.Printf("Starting command '%v'\n", m.cleanedCommand())

	if err := m.termClear(); err != nil {
		log.Println("Error sending clear to websocket:", err)
	}

	m.cmdMutex.Lock()
	m.cmdContext, m.cancelCommand = context.WithCancel(context.Background())
	m.cmd = exec.CommandContext(m.cmdContext, "sh", "-c", m.cleanedCommand())
	stdoutPipe, err := m.cmd.StdoutPipe()
	if err != nil {
		log.Printf("Error creating StdoutPipe for Cmd: %v\n", err)
		return
	}
	m.cmd.Stderr = m.cmd.Stdout
	m.cmdMutex.Unlock()

	runningButton(true).Render(w)

	go func() {
		if err := m.executeCommand(stdoutPipe); err != nil {
			log.Println(err.Error())
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

func (m *DemoManager) stopCurrentCommand() {
	m.cmdMutex.Lock()
	defer m.cmdMutex.Unlock()

	if m.cancelCommand != nil {
		m.cancelCommand()
		m.cancelCommand = nil
	}

	if m.cmd == nil {
		return
	}

	if m.cmd.Process != nil && (m.cmd.ProcessState == nil || !m.cmd.ProcessState.Exited()) {
		if err := m.cmd.Process.Signal(os.Interrupt); err != nil {
			log.Printf("Error stopping command: %v\n", err)
		}
	}
	m.cmd = nil

	log.Printf("Stopped command '%v'\n", m.cleanedCommand())
}

func (m *DemoManager) stopCommandHandler(w http.ResponseWriter, r *http.Request) {
	m.stopCurrentCommand()
	runningButton(false).Render(w)
}
