package server

import (
	"bufio"
	"context"
	"errors"
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

var commandRegex = regexp.MustCompile(`^\$\s*`)

var (
	ws          *websocket.Conn
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

func (m *DemoManager) indexHandler(w http.ResponseWriter, r *http.Request) {
	m.cmdMutex.Lock()
	isCmdRunning := m.cmd != nil
	m.cmdMutex.Unlock()

	w.WriteHeader(http.StatusOK)
	indexHTML(m.commands, m.cmdNumber, isCmdRunning).Render(w)
}

func (m *DemoManager) initHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: handle this better, js should try to reconect and refreshing shouldn't cause issues
	upgradeOnce.Do(func() {
		log.Println("Upgrading http to websocket")
		var err error
		ws, err = upgrader.Upgrade(w, r, nil) // upgrade HTTP to WebSocket
		if err != nil {
			log.Println("Error upgrading to websocket:", err)
			return
		}
	})
}

func (m *DemoManager) incPageHandler(w http.ResponseWriter, r *http.Request) {
	m.cmdNumberMutex.Lock()
	prevCommand := m.cmdNumber
	m.cmdNumber = min(len(m.commands)-1, m.cmdNumber+1)
	m.cmdNumberMutex.Unlock()

	m.cmdMutex.Lock()
	isCmdRunning := m.cmd != nil
	m.cmdMutex.Unlock()

	m.stopCurrentCommand()

	if prevCommand != m.cmdNumber {
		contentDiv(m.commands, m.cmdNumber, isCmdRunning).Render(w)
	} else {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err := ws.WriteMessage(websocket.TextMessage, []byte("\033[2J\033[H")); err != nil {
		log.Println("Error sending clear to websocket:", err)
	}
}

func (m *DemoManager) decPageHandler(w http.ResponseWriter, r *http.Request) {
	m.cmdNumberMutex.Lock()
	prevCommand := m.cmdNumber
	m.cmdNumber = max(0, m.cmdNumber-1)
	m.cmdNumberMutex.Unlock()

	m.cmdMutex.Lock()
	isCmdRunning := m.cmd != nil
	m.cmdMutex.Unlock()

	m.stopCurrentCommand()

	if prevCommand != m.cmdNumber {
		contentDiv(m.commands, m.cmdNumber, isCmdRunning).Render(w)
	} else {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err := ws.WriteMessage(websocket.TextMessage, []byte("\033[2J\033[H")); err != nil {
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

	m.cmdNumberMutex.Lock()
	prevCommand := m.cmdNumber
	m.cmdNumber = slideIndexVal
	m.cmdNumberMutex.Unlock()

	m.cmdMutex.Lock()
	isCmdRunning := m.cmd != nil
	m.cmdMutex.Unlock()

	if m.cmdNumber != prevCommand {
		m.stopCurrentCommand()
		contentDiv(m.commands, m.cmdNumber, isCmdRunning).Render(w)
	}

	if err := ws.WriteMessage(websocket.TextMessage, []byte("\033[2J\033[H")); err != nil {
		log.Println("Error sending clear to websocket:", err)
	}
}

func (m *DemoManager) waitForCommandExit() {
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

func (m *DemoManager) executeCommand(stdoutPipe io.Reader) {
	m.cmdMutex.Lock()
	if m.cmd == nil {
		return
	}
	if err := m.cmd.Start(); err != nil {
		log.Printf("Error starting command '%v': %v\n", commandRegex.ReplaceAllString(m.commands[m.cmdNumber], ""), err)
		return
	}
	m.cmdMutex.Unlock()

	go m.waitForCommandExit()

	reader := bufio.NewReader(stdoutPipe)
	buffer := make([]byte, terminalBufferSize)

	for {
		select {
		case <-m.cmdContext.Done():
			return // command context cancelled
		default:
			n, err := reader.Read(buffer)
			if err != nil {
				if err != io.EOF && !errors.Is(err, fs.ErrClosed) {
					log.Printf("Error reading from StdoutPipe: %v\n", err)
				}
				return // command complete
			}

			message := string(buffer[:n])
			message = strings.ReplaceAll(message, "\n", "\n\r")
			if err := ws.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
				log.Println("Error writing to websocket:", err)
				return
			}
		}
	}
}

func (m *DemoManager) executeCommandHandler(w http.ResponseWriter, r *http.Request) {
	m.stopCurrentCommand()
	log.Printf("Starting command '%v'\n", commandRegex.ReplaceAllString(m.commands[m.cmdNumber], ""))

	if err := ws.WriteMessage(websocket.TextMessage, []byte("\033[2J\033[H")); err != nil {
		log.Println("Error sending clear to websocket:", err)
	}

	m.cmdMutex.Lock()
	m.cmdContext, m.cancelCommand = context.WithCancel(context.Background())
	m.cmd = exec.CommandContext(m.cmdContext, "sh", "-c", commandRegex.ReplaceAllString(m.commands[m.cmdNumber], ""))
	stdoutPipe, err := m.cmd.StdoutPipe()
	if err != nil {
		log.Printf("Error creating StdoutPipe for Cmd: %v\n", err)
		return
	}
	m.cmd.Stderr = m.cmd.Stdout
	m.cmdMutex.Unlock()

	runningButton(true).Render(w)

	go m.executeCommand(stdoutPipe)
}

func (m *DemoManager) executeStatusHandler(w http.ResponseWriter, r *http.Request) {
	m.cmdMutex.Lock()
	isCmdRunning := m.cmd != nil
	m.cmdMutex.Unlock()

	if !isCmdRunning {
		runningButton(isCmdRunning).Render(w)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func (m *DemoManager) stopCommandHandler(w http.ResponseWriter, r *http.Request) {
	m.stopCurrentCommand()
	log.Printf("Stopped command '%v'\n", commandRegex.ReplaceAllString(m.commands[m.cmdNumber], ""))
	runningButton(false).Render(w)
}

func (m *DemoManager) stopCurrentCommand() {
	m.cmdMutex.Lock()
	defer m.cmdMutex.Unlock()

	if m.cancelCommand != nil {
		m.cancelCommand()
		m.cancelCommand = nil
	}

	if m.cmd != nil && m.cmd.Process != nil && (m.cmd.ProcessState == nil || !m.cmd.ProcessState.Exited()) {
		if err := m.cmd.Process.Signal(os.Interrupt); err != nil {
			log.Printf("Error stopping command: %v\n", err)
		}
	}
	m.cmd = nil
}
