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

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"

	"github.com/joshjennings98/backend-demo/server/v2/types"
)

const (
	EndpointIndex         = "GET /presentation"
	EndpointInit          = "GET /init"
	EndpointSlideByIndex  = "GET /slides/{id}/"
	EndpointSlideByQuery  = "GET /slides/"
	EndpointCommandStart  = "POST /commands/{id}/start"
	EndpointCommandStatus = "GET /commands/{id}/status"
	EndpointCommandStop   = "POST /commands/{id}/stop"
)

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
	if strings.HasPrefix(content, "$ ") {
		return strings.TrimPrefix(content, "$ ")
	}

	var html strings.Builder
	if err := md2html.Convert([]byte(content), &html); err == nil {
		return html.String()
	}

	return content
}

type server struct {
	preCommands    []string
	slides         []types.Slide
	commandsFile   string
	commandManager ICommandManager
	logger         *slog.Logger
}

func (s *server) GetPreCommands() []string {
	return s.preCommands
}

func (s *server) GetSlide(idx int) types.Slide {
	return s.slides[idx]
}

func (s *server) GetSlideCount() int {
	return len(s.slides)
}

func (s *server) ParsePreCommands(contents []string) (err error) {
	s.preCommands = slices.DeleteFunc(contents, func(s string) bool { return whiteSpaceRegex.MatchString(s) })
	return
}

func (s *server) ParseSlide(content string) {
	var slideType types.SlideType

	switch {
	case whiteSpaceRegex.MatchString(content):
		return
	case strings.HasPrefix(content, "```"):
		slideType = types.SlideTypeCodeblock
	case strings.HasPrefix(content, "$ "):
		slideType = types.SlideTypeCommand
	default:
		slideType = types.SlideTypePlain
	}

	s.slides = append(s.slides, types.Slide{
		ID:        len(s.slides),
		Content:   parseSlide(content),
		SlideType: slideType,
	})
}

func (s *server) Initialise(ctx context.Context) (err error) {
	for _, cmd := range s.GetPreCommands() {
		cmd = strings.TrimSpace(cmd)

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if strings.HasPrefix(cmd, "#") {
				continue
			}

			cmd := varCommandWithComment.ReplaceAllStringFunc(cmd, func(match string) string {
				submatches := varCommandWithComment.FindStringSubmatch(match)
				return strings.TrimSpace(fmt.Sprintf("%v", submatches[1]))
			})

			if whiteSpaceRegex.MatchString(cmd) {
				continue
			}

			s.logger.Info("running pre-command", "command", cmd)

			switch {
			case varCommandPattern.MatchString(cmd):
				if err := setCommandToVar(ctx, cmd); err != nil {
					err = fmt.Errorf("could not set command to var '%v': %v", cmd, err.Error())
					return err
				}
			case varPattern.MatchString(cmd):
				if err := setVar(cmd); err != nil {
					err = fmt.Errorf("could set variable '%v': %v", cmd, err.Error())
					return err
				}
			default:
				if err := runCommand(ctx, cmd); err != nil {
					err = fmt.Errorf("could not execute pre-command '%v': %v", cmd, err.Error())
					return err
				}
			}
		}
	}

	return
}

func (s *server) ParseSlides(contents []string) (err error) {
	for i := range contents {
		line := contents[i]
		trimmed := strings.TrimSpace(line)

		if trimmed != "" {
			s.ParseSlide(trimmed)
		}
	}

	return
}

func (s *server) SplitContent(commandsFile string) (preCommands, slideContent []string, err error) {
	contents, err := os.ReadFile(commandsFile)
	if err != nil {
		return
	}

	var preCommandsStr, slideContentStr string
	preCommandsSplit := regexp.MustCompile(`\n*(\-+)\n+`).FindAllIndex(contents, -1)

	switch {
	case len(preCommandsSplit) == 0:
		slideContentStr = string(contents)
	case len(preCommandsSplit) > 1:
		err = errors.New("pre commands must be separated from the rest of the presentation by '---' but more than one '---' was found")
		return
	default:
		preCommandsStr = string(contents[:preCommandsSplit[0][0]])
		slideContentStr = string(contents[preCommandsSplit[0][1]:])
		preCommands = strings.Split(preCommandsStr, "\n")
	}

	slideContent = strings.Split(regexp.MustCompile(`\\\s*\n`).ReplaceAllString(slideContentStr, ""), "\n\n")
	slideContent = slices.DeleteFunc(slideContent, func(s string) bool { return whiteSpaceRegex.MatchString(s) })
	return
}

func NewServer(logger *slog.Logger, commandsFile string) (s IPresentationServer, err error) {
	s = &server{
		logger:         logger,
		commandsFile:   commandsFile,
		commandManager: newCommandManager(logger),
	}

	preComamnds, slideContent, err := s.SplitContent(commandsFile)
	if err != nil {
		return
	}

	err = s.ParsePreCommands(preComamnds)
	if err != nil {
		err = errors.New("could not parse pre-commands")
		return
	}

	err = s.ParseSlides(slideContent)
	if err != nil {
		err = errors.New("could not parse slides")
		return
	}

	return
}

//go:embed static
var staticFS embed.FS

func (s *server) Start(ctx context.Context) (err error) {
	err = s.Initialise(ctx)
	if err != nil {
		s.logger.Error("could not initialise server", "error", err.Error())
		return err
	}

	mux := http.NewServeMux()

	mux.HandleFunc(EndpointIndex, s.HandlerIndex)
	mux.HandleFunc(EndpointInit, s.HandlerInit)
	mux.HandleFunc(EndpointSlideByIndex, s.HandlerSlideByIndex)
	mux.HandleFunc(EndpointSlideByQuery, s.HandlerSlideByQuery)
	mux.HandleFunc(EndpointCommandStart, s.HandlerCommandStart)
	mux.HandleFunc(EndpointCommandStatus, s.HandlerCommandStatus)
	mux.HandleFunc(EndpointCommandStop, s.HandlerCommandStop)

	mux.HandleFunc("/static/", http.FileServerFS(staticFS).ServeHTTP)
	mux.HandleFunc("/", http.FileServer(http.Dir(filepath.Dir(s.commandsFile))).ServeHTTP)

	s.logger.Info("server is running", "host", "http://localhost:8080/presentation")

	return http.ListenAndServe("localhost:8080", mux) //nolint:gosec
}
