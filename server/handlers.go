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
	"github.com/maragudk/gomponents"
	"github.com/maragudk/gomponents/html"
)

var (
	commands       = []string{}
	cmd            *exec.Cmd
	cmdMutex       sync.Mutex
	cmdNumber      int
	cmdNumberMutex sync.Mutex
	cmdContext     context.Context
	cancelCommand  func()
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

func indexHandler(w http.ResponseWriter, r *http.Request) {
	page := html.HTML(
		html.Head(
			html.TitleEl(gomponents.Text("Backend Demo Tool")),
			html.Script(html.Src("static/main.js")),
			html.Script(html.Src("https://unpkg.com/htmx.org")),
			html.Script(html.Src("https://cdn.jsdelivr.net/npm/xterm/lib/xterm.js")),
			html.Link(html.Rel("stylesheet"), html.Href("static/main.css")),
			html.Link(html.Rel("stylesheet"), html.Href("https://cdn.jsdelivr.net/npm/xterm/css/xterm.css")),
		),
		html.Body(
			contentDiv(),
		),
	)

	w.WriteHeader(http.StatusOK)
	page.Render(w)
}

func initHandler(w http.ResponseWriter, r *http.Request) {
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

func incPageHandler(w http.ResponseWriter, r *http.Request) {
	cmdNumberMutex.Lock()
	prevCommand := cmdNumber
	cmdNumber = min(len(commands)-1, cmdNumber+1)
	cmdNumberMutex.Unlock()

	stopCurrentCommand()

	if prevCommand != cmdNumber {
		contentDiv().Render(w)
	} else {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err := ws.WriteMessage(websocket.TextMessage, []byte("\033[2J\033[H")); err != nil {
		log.Println("Error sending clear to websocket:", err)
	}
}

func decPageHandler(w http.ResponseWriter, r *http.Request) {
	cmdNumberMutex.Lock()
	prevCommand := cmdNumber
	cmdNumber = max(0, cmdNumber-1)
	cmdNumberMutex.Unlock()

	stopCurrentCommand()

	if prevCommand != cmdNumber {
		contentDiv().Render(w)
	} else {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err := ws.WriteMessage(websocket.TextMessage, []byte("\033[2J\033[H")); err != nil {
		log.Println("Error sending clear to websocket:", err)
	}
}

func setPageHandler(w http.ResponseWriter, r *http.Request) {
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

	cmdNumberMutex.Lock()
	prevCommand := cmdNumber
	cmdNumber = slideIndexVal
	cmdNumberMutex.Unlock()

	if cmdNumber != prevCommand {
		stopCurrentCommand()
		contentDiv().Render(w)
	}

	if err := ws.WriteMessage(websocket.TextMessage, []byte("\033[2J\033[H")); err != nil {
		log.Println("Error sending clear to websocket:", err)
	}
}

func waitForCommandExit() {
	var exitErr *exec.ExitError
	if err := cmd.Wait(); errors.As(err, &exitErr) {
		return // command exited
	}
	select {
	case <-cmdContext.Done():
		return // already cancelled
	default:
		stopCurrentCommand()
		return
	}
}

func executeCommand(stdoutPipe io.Reader) {
	cmdMutex.Lock()
	if cmd == nil {
		return
	}
	if err := cmd.Start(); err != nil {
		log.Printf("Error starting command '%v': %v\n", commandRegex.ReplaceAllString(commands[cmdNumber], ""), err)
		return
	}
	cmdMutex.Unlock()

	go waitForCommandExit()

	reader := bufio.NewReader(stdoutPipe)
	buffer := make([]byte, terminalBufferSize)

	for {
		select {
		case <-cmdContext.Done():
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

func executeCommandHandler(w http.ResponseWriter, r *http.Request) {
	stopCurrentCommand()
	log.Printf("Starting command '%v'\n", commandRegex.ReplaceAllString(commands[cmdNumber], ""))

	if err := ws.WriteMessage(websocket.TextMessage, []byte("\033[2J\033[H")); err != nil {
		log.Println("Error sending clear to websocket:", err)
	}

	cmdMutex.Lock()
	cmdContext, cancelCommand = context.WithCancel(context.Background())
	cmd = exec.CommandContext(cmdContext, "sh", "-c", commandRegex.ReplaceAllString(commands[cmdNumber], ""))
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("Error creating StdoutPipe for Cmd: %v\n", err)
		return
	}
	cmd.Stderr = cmd.Stdout
	cmdMutex.Unlock()

	runningButton().Render(w)

	go executeCommand(stdoutPipe)
}

func executeStatusHandler(w http.ResponseWriter, r *http.Request) {
	cmdMutex.Lock()
	isCmdRunning := cmd != nil
	cmdMutex.Unlock()

	if !isCmdRunning {
		runningButton().Render(w)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func stopCommandHandler(w http.ResponseWriter, r *http.Request) {
	stopCurrentCommand()
	log.Printf("Stopped command '%v'\n", commandRegex.ReplaceAllString(commands[cmdNumber], ""))
	runningButton().Render(w)
}

func stopCurrentCommand() {
	cmdMutex.Lock()
	defer cmdMutex.Unlock()

	if cancelCommand != nil {
		cancelCommand()
		cancelCommand = nil
	}

	if cmd != nil && cmd.Process != nil && (cmd.ProcessState == nil || !cmd.ProcessState.Exited()) {
		if err := cmd.Process.Signal(os.Interrupt); err != nil {
			log.Printf("Error stopping command: %v\n", err)
		}
	}
	cmd = nil
}
