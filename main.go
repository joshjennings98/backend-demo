package main

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

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/maragudk/gomponents"
	"github.com/maragudk/gomponents/html"
)

var (
	command int
	mu      sync.Mutex
)

func main() {
	contents, err := os.ReadFile("commands.txt")
	if err != nil {
		log.Fatal(err)
	}

	split := strings.Split(regexp.MustCompile(`\\\s*\n`).ReplaceAllString(string(contents), ""), "\n")
	for i := range split {
		if strings.TrimSpace(split[i]) != "" {
			commands = append(commands, split[i])
		}
	}

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	r := mux.NewRouter()
	r.HandleFunc("/", handleIndex)
	r.HandleFunc("/init", handleWebsocket)
	r.HandleFunc("/inc-page", handleIncrement)
	r.HandleFunc("/dec-page", handleDecrement)
	r.HandleFunc("/set-page", handleSetPage)
	r.HandleFunc("/execute", handleExecute).Methods(http.MethodPost)
	r.HandleFunc("/execute", handleCommandStatus).Methods(http.MethodGet)
	r.HandleFunc("/stop", handleStop)

	http.Handle("/", r)
	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

// TODO: simplify this so that less is updated
func contentScreen() gomponents.Node {
	isCommand := strings.HasPrefix(commands[command], "$")
	return html.Div(
		gomponents.Attr("id", "command"),
		html.Div(
			gomponents.Attr("id", "controls"),
			createDropdown(command),
			html.FormEl(
				gomponents.Attr("class", "control"),
				gomponents.Attr("method", "post"),
				gomponents.Attr("hx-post", "/dec-page"),
				gomponents.Attr("hx-swap", "outerHTML"),
				gomponents.Attr("hx-target", "#command"),
				gomponents.Attr("hx-trigger", "click, keyup[key=='ArrowLeft'] from:body"),
				html.Button(gomponents.Text("prev")),
			),
			html.FormEl(
				gomponents.Attr("class", "control"),
				gomponents.Attr("method", "post"),
				gomponents.Attr("hx-post", "/inc-page"),
				gomponents.Attr("hx-swap", "outerHTML"),
				gomponents.Attr("hx-target", "#command"),
				gomponents.Attr("hx-trigger", "click, keyup[key=='ArrowRight'] from:body"),
				html.Button(gomponents.Text("next")),
			),
		),
		html.Div(
			html.Div(
				gomponents.If(
					isCommand,
					html.Class("command-string"),
				),
				gomponents.If(
					!isCommand,
					html.Class("text-string"),
				),
				gomponents.Text(commandRegex.ReplaceAllString(commands[command], "")),
			),
		),
		gomponents.If(isCommand,
			html.Div(
				html.Div(
					gomponents.Attr("id", "terminal"),
					gomponents.Attr("hx-preserve", "true"),
				),
				runButton(),
			),
		),
		gomponents.If(!isCommand,
			html.Div(
				gomponents.Attr("hidden", "true"),
				html.Div(
					gomponents.Attr("id", "terminal"),
					gomponents.Attr("hx-preserve", "true"),
				),
				runButton(),
			),
		),
	)
}

func createDropdown(selected int) gomponents.Node {
	var options []gomponents.Node
	for i := range commands {
		var optionAttributes []gomponents.Node
		optionAttributes = append(optionAttributes, html.Value(fmt.Sprintf("%d", i)))

		if i == selected {
			optionAttributes = append(optionAttributes, html.Selected())
		}

		options = append(options, html.Option(
			gomponents.Group(optionAttributes),
			gomponents.Text(fmt.Sprintf("Slide %d/%d", i+1, len(commands))),
		))
	}

	return html.Select(
		gomponents.Attr("hx-post", "/set-page"),
		gomponents.Attr("hx-trigger", "change"),
		gomponents.Attr("hx-target", "#command"),
		gomponents.Attr("name", "slideIndex"),
		gomponents.Group(options),
	)
}

func handleSetPage(w http.ResponseWriter, r *http.Request) {
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

	mu.Lock()
	prevCommand := command
	command = slideIndexVal
	mu.Unlock()
	if command != prevCommand {
		stopCurrentCommand()
		contentScreen().Render(w)
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	// Render the entire page
	page := html.HTML(
		html.Head(
			html.TitleEl(gomponents.Text("Backend Demo Tool")),
			html.Script(gomponents.Attr("src", "https://unpkg.com/htmx.org")),
			html.Script(gomponents.Attr("src", "https://cdn.jsdelivr.net/npm/xterm/lib/xterm.js")),
			html.Script(gomponents.Attr("src", "static/main.js")),
			html.Link(html.Rel("stylesheet"), html.Href("static/main.css")),
			html.Link(html.Rel("stylesheet"), html.Href("https://cdn.jsdelivr.net/npm/xterm/css/xterm.css")),
		),
		html.Body(
			contentScreen(),
		),
	)

	w.WriteHeader(http.StatusOK)
	page.Render(w)
}

func handleIncrement(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	prevCommand := command
	command = min(len(commands)-1, command+1)
	mu.Unlock()
	stopCurrentCommand()

	if prevCommand != command {
		contentScreen().Render(w)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}

	if err := conn.WriteMessage(websocket.TextMessage, []byte("\033[2J\033[H")); err != nil {
		log.Println("Error sending clear to websocket:", err)
	}
}

func handleStop(w http.ResponseWriter, r *http.Request) {
	stopCurrentCommand()
	runButton().Render(w)
}

func handleDecrement(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	prevCommand := command
	command = max(0, command-1)
	mu.Unlock()
	stopCurrentCommand()

	if prevCommand != command {
		contentScreen().Render(w)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}

	if err := conn.WriteMessage(websocket.TextMessage, []byte("\033[2J\033[H")); err != nil {
		log.Println("Error sending clear to websocket:", err)
	}
}

const bufferSize = 1024

var commands = []string{}

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

var commandRegex = regexp.MustCompile(`^\$\s*`)

func handleExecute(w http.ResponseWriter, r *http.Request) {
	stopCurrentCommand()
	log.Println("starting command")

	if err := conn.WriteMessage(websocket.TextMessage, []byte("\033[2J\033[H")); err != nil {
		log.Println("Error sending clear to websocket:", err)
	}

	cmdMutex.Lock()
	cmdContext, cancelCommand = context.WithCancel(context.Background())
	cmd = exec.CommandContext(cmdContext, "sh", "-c", commandRegex.ReplaceAllString(commands[command], ""))
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
		if cmd == nil {
			return
		}
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

func newControlButton(buttonType string) gomponents.Node {
	return gomponents.Group([]gomponents.Node{ // TODO: make these class based for the targets?
		html.FormEl(
			gomponents.Attr("id", fmt.Sprintf("%v-button", buttonType)),
			gomponents.Attr("method", "post"),
			gomponents.Attr("hx-post", fmt.Sprintf("/%v", buttonType)),
			gomponents.Attr("hx-target", fmt.Sprintf("#%v-button", buttonType)),
			html.Button(gomponents.Text(strings.Title(buttonType))),
		),
		html.FormEl(
			gomponents.Attr("id", fmt.Sprintf("%v-button", buttonType)),
			gomponents.Attr("method", "post"),
			gomponents.Attr("hx-post", fmt.Sprintf("/%v", buttonType)),
			gomponents.Attr("hx-target", fmt.Sprintf("#%v-button", buttonType)),
			gomponents.Attr("hx-trigger", "keyup[key==' '] from:body"),
			gomponents.Attr("hidden", "true"),
		),
	})
}

func stopButton() gomponents.Node {
	return newControlButton("stop")
}

func executeButton() gomponents.Node {
	return newControlButton("execute")
}
