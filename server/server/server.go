package server

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"

	"github.com/joshjennings98/backend-demo/server/v2/types"
)

//go:embed static
var staticFS embed.FS

const (
	EndpointIndex         = "GET  /presentation"
	EndpointWebSocket     = "GET  /ws"
	EndpointSlideByIndex  = "GET  /slides/{id}/"
	EndpointSlideByQuery  = "GET  /slides/"
	EndpointCommandStart  = "POST /commands/{id}/start"
	EndpointCommandStatus = "GET  /commands/{id}/status"
	EndpointCommandStop   = "POST /commands/{id}/stop"
)

var ErrSlideIndexOutOfBounds = errors.New("slide index out of bounds")

var (
	whiteSpaceRegex = regexp.MustCompile(`^[\s\n\r]*$`)
	md2html         = goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
			html.WithUnsafe(),
		),
	)
)

func parseSlide(content string) (output string) {
	var html strings.Builder
	err := md2html.Convert([]byte(content), &html)
	if err == nil {
		output = html.String()
	} else {
		output = content
	}

	return
}

// parseCommandSlide parses a multi-line command block.
// Returns displayContent (visible lines) and executeContent (all commands to run).
// Lines with $! are visible, the last $ line is always visible.
// All $ and $! lines are executed.
func parseCommandSlide(content string) (displayContent []string, executeContent []string) {
	lines := strings.Split(content, "\n")
	lines = slices.DeleteFunc(lines, func(line string) bool {
		return strings.TrimSpace(line) == ""
	})

	for i, line := range lines {
		switch {
		case strings.HasPrefix(line, "$! "):
			cmd := strings.TrimPrefix(line, "$! ")
			displayContent = append(displayContent, cmd)
			executeContent = append(executeContent, cmd)
		case strings.HasPrefix(line, "$ "):
			cmd := strings.TrimPrefix(line, "$ ")
			executeContent = append(executeContent, cmd)
			if i == len(lines)-1 { // the last $ line is always visible
				displayContent = append(displayContent, cmd)
			}
		}
	}

	return
}

type server struct {
	port           int
	slides         []types.Slide
	commandsFile   string
	commandManager ICommandManager
	logger         *slog.Logger
}

func (s *server) GetSlide(idx int) (slide types.Slide, err error) {
	if idx < 0 || idx >= len(s.slides) {
		err = ErrSlideIndexOutOfBounds
		return
	}

	slide = s.slides[idx]
	return
}

func (s *server) GetSlideCount() int {
	return len(s.slides)
}

func isCommand(content string) (isCommand bool) {
	for line := range strings.SplitSeq(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "$ ") || strings.HasPrefix(line, "$! ") {
			isCommand = true
			return
		}
	}

	return
}

func (s *server) ParseSlide(content string) {
	if whiteSpaceRegex.MatchString(content) {
		return
	}

	slide := types.Slide{
		ID: len(s.slides),
	}

	switch {
	case strings.HasPrefix(content, "```"):
		slide.SlideType = types.SlideTypeCodeblock
		slide.Content = parseSlide(content)
	case isCommand(content):
		slide.SlideType = types.SlideTypeCommand
		displayContent, executeContent := parseCommandSlide(content)
		slide.Content = strings.Join(displayContent, "\n")
		slide.ExecuteContent = executeContent
	default:
		slide.SlideType = types.SlideTypePlain
		slide.Content = parseSlide(content)
	}

	s.slides = append(s.slides, slide)
}

func (s *server) ParseSlides(contents []string) {
	for _, line := range contents {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			s.ParseSlide(trimmed)
		}
	}
}

func (s *server) SplitContent(commandsFile string) (slideContent []string, err error) {
	contents, err := os.ReadFile(commandsFile)
	if err != nil {
		err = fmt.Errorf("could not read content from '%v': %w", commandsFile, err)
		return
	}

	// handle line continuations (\ at end of line)
	contentStr := regexp.MustCompile(`\\\s*\n`).ReplaceAllString(string(contents), "")

	// split by double newlines to get slide blocks
	slideContent = strings.Split(contentStr, "\n\n")
	slideContent = slices.DeleteFunc(slideContent, func(s string) bool { return whiteSpaceRegex.MatchString(s) })
	return
}

func NewServer(logger *slog.Logger, port int, commandsFile string) (s IPresentationServer, err error) {
	if port == 0 {
		port = 8080
	}

	s = &server{
		port:           port,
		logger:         logger,
		commandsFile:   commandsFile,
		commandManager: newCommandManager(logger),
	}

	content, err := s.SplitContent(commandsFile)
	if err != nil {
		err = fmt.Errorf("could not split content of file '%v': %w", commandsFile, err)
		return
	}

	s.ParseSlides(content)
	return
}

func (s *server) Start(ctx context.Context) error {
	mux := http.NewServeMux()

	mux.HandleFunc(EndpointIndex, s.HandlerIndex)
	mux.HandleFunc(EndpointWebSocket, s.HandlerWebSocket)
	mux.HandleFunc(EndpointSlideByIndex, s.HandlerSlideByIndex)
	mux.HandleFunc(EndpointSlideByQuery, s.HandlerSlideByQuery)
	mux.HandleFunc(EndpointCommandStart, s.HandlerCommandStart)
	mux.HandleFunc(EndpointCommandStatus, s.HandlerCommandStatus)
	mux.HandleFunc(EndpointCommandStop, s.HandlerCommandStop)

	mux.HandleFunc("/static/", http.FileServerFS(staticFS).ServeHTTP)
	mux.HandleFunc("/", http.FileServer(http.Dir(filepath.Dir(s.commandsFile))).ServeHTTP)

	s.logger.Info("server is running", "host", fmt.Sprintf("http://localhost:%v/presentation", s.port))

	server := &http.Server{
		Addr:              fmt.Sprintf("localhost:%v", s.port),
		Handler:           mux,
		ReadHeaderTimeout: 3 * time.Second, // https://deepsource.com/directory/go/issues/GO-S2112
	}

	go func() {
		<-ctx.Done()
		_ = server.Shutdown(context.Background())
	}()

	return server.ListenAndServe()
}
