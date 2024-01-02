package main

import (
	"bufio"
	"context"
    "errors"
	"fmt"
    "io/fs"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
    "github.com/gorilla/mux"
	"github.com/maragudk/gomponents"
	"github.com/maragudk/gomponents/html"
)

var (
	counter int
	mu      sync.Mutex
)

func main() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
    
    r := mux.NewRouter()
	r.HandleFunc("/", handleIndex)
    r.HandleFunc("/init", handleWebsocket)
	r.HandleFunc("/increment", handleIncrement)
	r.HandleFunc("/decrement", handleDecrement)
	r.HandleFunc("/execute", handleExecute).Methods(http.MethodPost)
	r.HandleFunc("/execute", handleCommandStatus).Methods(http.MethodGet)
	r.HandleFunc("/stop", handleStop)

    http.Handle("/", r)
	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	// Render the entire page
	page := html.HTML(
		html.Head(
			html.TitleEl(gomponents.Text("Backend Demo Tool")),
			html.Script(gomponents.Attr("src", "https://unpkg.com/htmx.org")),
			html.Script(gomponents.Attr("src", "https://cdn.jsdelivr.net/npm/xterm/lib/xterm.js")),
			html.Script(gomponents.Attr("src", "static/main.js")),
			html.Link(html.Rel("stylesheet"), html.Href("https://cdn.jsdelivr.net/npm/xterm/css/xterm.css")),
		),
		html.Body(
			html.Div(
				gomponents.Attr("id", "counter-display"),
				commandDisplay(),
				runButton(),
			),
			html.Div(
				gomponents.Attr("id", "terminal"),
			),
			html.Div(
				html.FormEl(
					gomponents.Attr("method", "post"),
					gomponents.Attr("hx-post", "/decrement"),
					gomponents.Attr("hx-target", "#counter-display"),
					html.Button(gomponents.Text("prev")),
				),
				html.FormEl(
					gomponents.Attr("method", "post"),
					gomponents.Attr("hx-post", "/increment"),
					gomponents.Attr("hx-target", "#counter-display"),
					html.Button(gomponents.Text("next")),
				),
			),
		),
	)

	w.WriteHeader(http.StatusOK)
	page.Render(w)
}

func handleIncrement(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	counter = min(len(commands), counter+1)
	mu.Unlock()
	stopCurrentCommand()
	commandDisplay().Render(w)
	runButton().Render(w)
}

func handleStop(w http.ResponseWriter, r *http.Request) {
	stopCurrentCommand()
	runButton().Render(w)
}

func handleDecrement(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	counter = max(0, counter-1)
	mu.Unlock()
	stopCurrentCommand()
	commandDisplay().Render(w)
	runButton().Render(w)
}

const bufferSize = 1024

var commands = []string{
	"echo aaa && sleep 2 && echo bbb",
	"watch -n 1 date",
	"ls /",
	// "",
	"echo {} | jq",
	"adsadads",
	"ls -R /",
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  bufferSize,
	WriteBufferSize: bufferSize,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow connections from any origin
	},
}

var (
	conn          *websocket.Conn
	upgradeOnce   sync.Once
	cmd           *exec.Cmd
	cmdMutex      sync.Mutex
	cmdContext    context.Context
	cancelCommand func()
)

func handleWebsocket(w http.ResponseWriter, r *http.Request) {
    // TODO: handle this better, js should try to reconect and refreshing shouldn't cause issues
	upgradeOnce.Do(func() {
        log.Println("upgrading http to websocket")
        var err error
        conn, err = upgrader.Upgrade(w, r, nil) // upgrade HTTP to WebSocket
        if err != nil {
            log.Println("Error upgrading to websocket:", err)
            return
        }
    })
}

func handleExecute(w http.ResponseWriter, r *http.Request) {
	stopCurrentCommand()
	log.Println("starting command")

	if err := conn.WriteMessage(websocket.TextMessage, []byte("\033[2J\033[H")); err != nil {
		log.Println("Error sending clear to websocket:", err)
	}

	cmdMutex.Lock()
	cmdContext, cancelCommand = context.WithCancel(context.Background())
	cmd = exec.CommandContext(cmdContext, "sh", "-c", commands[counter])
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("Error creating StdoutPipe for Cmd: %v\n", err)
		return
	}
	cmd.Stderr = cmd.Stdout
	cmdMutex.Unlock()

	lastKnownState = true
	runButton().Render(w)

	go func() {

		cmdMutex.Lock()
		if err := cmd.Start(); err != nil {
			log.Printf("Error starting Cmd: %v\n", err)
			return
		}
		cmdMutex.Unlock()
        
        go func() {
            var exitErr *exec.ExitError
            if err := cmd.Wait(); errors.As(err, &exitErr) {
                log.Println("command exited")
                return
            }
            select {
            case <-cmdContext.Done():
                log.Println("already cancelled")
                return
            default:
                stopCurrentCommand()
                return
            }
        }()

		reader := bufio.NewReader(stdoutPipe)
		buffer := make([]byte, bufferSize)

		for {
			select {
			case <-cmdContext.Done():
                log.Println("cancelled")
				return // Exit the goroutine if the context is canceled
			default:
				n, err := reader.Read(buffer)
				if err != nil {
					if err != io.EOF && !errors.Is(err, fs.ErrClosed) {
						log.Printf("Error reading from StdoutPipe: %v\n", err)
					}
					return // Exit the goroutine if command complete
				}

				message := string(buffer[:n])
				message = strings.ReplaceAll(message, "\n", "\n\r")
				if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
					log.Println("Error writing to websocket:", err)
					return
				}
			}
		}
	}()
}

func stopCurrentCommand() {
    log.Println("stopping command")
	cmdMutex.Lock()
    log.Println("gained mutex")
	defer cmdMutex.Unlock()

	if cancelCommand != nil {
		log.Println("cancelling command")
		cancelCommand() // Cancel the context, which stops the command and exits the for loop
		cancelCommand = nil
	}

	if cmd != nil && cmd.Process != nil && (cmd.ProcessState == nil || !cmd.ProcessState.Exited()) {
		if err := cmd.Process.Signal(os.Interrupt); err != nil {
			log.Printf("Error stopping command: %v\n", err)
		}
	}
	cmd = nil
    log.Println("stopped command")
}

func commandDisplay() gomponents.Node {
	return html.Div(
		html.P(gomponents.Text(fmt.Sprint(commands[counter]))),
	)
}

var (
	lastKnownState = false
)

func handleCommandStatus(w http.ResponseWriter, r *http.Request) {
	cmdMutex.Lock()
	isCmdRunning := cmd != nil
	cmdMutex.Unlock()

	if !isCmdRunning {
		runButton().Render(w)
	} else {
		w.WriteHeader(http.StatusNoContent) // Send a 204 No Content to avoid unnecessary updates
	}
}

func runButton() gomponents.Node {
	cmdMutex.Lock()
	isCmdRunning := cmd != nil
	cmdMutex.Unlock()

	if isCmdRunning {
		return html.Div(
			gomponents.Attr("hx-get", "/execute"),
			gomponents.Attr("hx-trigger", "every 100ms"),
			gomponents.Attr("hx-target", "#execute-button"),
			stopButton(),
		)
	}

	return html.Div(
		executeButton(),
	)
}

func newButton(buttonType string) gomponents.Node {
	return html.FormEl(
		gomponents.Attr("id", fmt.Sprintf("%v-button", buttonType)),
		gomponents.Attr("method", "post"),
		gomponents.Attr("hx-post", fmt.Sprintf("/%v", buttonType)),
		gomponents.Attr("hx-target", fmt.Sprintf("#%v-button", buttonType)),
		html.Button(gomponents.Text(strings.Title(buttonType))),
	)
}

func stopButton() gomponents.Node {
    return newButton("stop")
}

func executeButton() gomponents.Node {
    return newButton("execute")
}
