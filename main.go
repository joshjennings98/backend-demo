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
	hx "github.com/maragudk/gomponents-htmx"
	"github.com/maragudk/gomponents/html"
)

var (
	commands       = []string{}
	cmd            *exec.Cmd
	cmdMutex       sync.Mutex
	cmdNumber      int
	cmdNumberMutex sync.Mutex
	lastKnownState = false
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
	r.HandleFunc(indexEndpoint, indexHandler)
	r.HandleFunc(initEndpoint, initHandler)
	r.HandleFunc(incPageEndpoint, incPageHandler)
	r.HandleFunc(decPageEndpoint, decPageHandler)
	r.HandleFunc(setPageEndpoint, setPageHandler)
	r.HandleFunc(executeEndpoint, executeCommandHandler).
		Methods(http.MethodPost)
	r.HandleFunc(executeEndpoint, executeStatusHandler).
		Methods(http.MethodGet)
	r.HandleFunc(stopEndpoint, stopCommandHandler)

	http.Handle(indexEndpoint, r)
	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func gomponentsIfElse(condition bool, ifBranch, elseBranch gomponents.Node) gomponents.Node {
	if condition {
		return ifBranch
	}
	return elseBranch
}

func contentDiv() gomponents.Node {
	isCommand := strings.HasPrefix(commands[cmdNumber], "$")
	return html.Div(
		html.ID("command"),
		html.Div(
			html.ID("controls"),
			slideSelect(cmdNumber),
			html.FormEl(
				html.Class("control"),
				hx.Post(decPageEndpoint),
				hx.Swap("outerHTML"),
				hx.Target("#command"),
				hx.Trigger("click, keyup[key=='ArrowLeft'] from:body"),
				html.Button(gomponents.Text("prev")),
			),
			html.FormEl(
				html.Class("control"),
				hx.Post(incPageEndpoint),
				hx.Swap("outerHTML"),
				hx.Target("#command"),
				hx.Trigger("click, keyup[key=='ArrowRight'] from:body"),
				html.Button(gomponents.Text("next")),
			),
		),
		html.Div(
			html.Div(
				gomponentsIfElse(
					isCommand,
					html.Class("command-string"),
					html.Class("text-string"),
				),
				gomponents.Text(commandRegex.ReplaceAllString(commands[cmdNumber], "")),
			),
		),
		html.Div(
			gomponents.If(
				!isCommand,
				gomponents.Attr("hidden", "true"),
			),
			html.Div(
				html.ID("terminal"),
				hx.Preserve("true"),
			),
			runningButton(),
		),
	)
}

func slideSelect(selected int) gomponents.Node {
	var options []gomponents.Node
	for i := range commands {
		options = append(options, html.Option(
			gomponents.Group([]gomponents.Node{
				html.Value(fmt.Sprint(i)),
				gomponents.If(
					i == selected,
					html.Selected(),
				),
			}),
			gomponents.Text(fmt.Sprintf("Slide %d/%d", i+1, len(commands))),
		))
	}

	return html.Select(
		hx.Post(setPageEndpoint),
		hx.Trigger("change"),
		hx.Target("#command"),
		html.Name("slideIndex"),
		gomponents.Group(options),
	)
}

func runningButton() gomponents.Node {
	cmdMutex.Lock()
	isCmdRunning := cmd != nil
	cmdMutex.Unlock()

	if isCmdRunning {
		return html.Div(
			hx.Get(executeEndpoint),
			hx.Trigger("every 100ms"),
			hx.Target("#execute-button"),
			stopButton(),
		)
	}

	return html.Div(
		executeButton(),
	)
}

func controlButton(buttonType string) gomponents.Node {
	return gomponents.Group([]gomponents.Node{
		html.FormEl(
			html.ID(fmt.Sprintf("%v-button", buttonType)),
			hx.Post(fmt.Sprintf("/%v", buttonType)),
			hx.Target(fmt.Sprintf("#%v-button", buttonType)),
			html.Button(gomponents.Text(buttonType)),
		),
		html.FormEl(
			hx.Post(fmt.Sprintf("/%v", buttonType)),
			hx.Target(fmt.Sprintf("#%v-button", buttonType)),
			hx.Trigger("keyup[key==' '] from:body"),
			gomponents.Attr("hidden", "true"),
		),
	})
}

func stopButton() gomponents.Node {
	return controlButton(stopEndpoint[1:])
}

func executeButton() gomponents.Node {
	return controlButton(executeEndpoint[1:])
}

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

	lastKnownState = true
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
