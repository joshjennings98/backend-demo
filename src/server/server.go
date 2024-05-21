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
	hrRegex                = regexp.MustCompile(`^\-+`)
	plainSlideReplacements = []struct {
		regex       *regexp.Regexp
		parseResult func(matches []string) string
	}{

		{
			regexp.MustCompile(`^!\[([^\]]*)\]\(([^)]+)\)$`), // only allow image only slides
			func(matches []string) string {
				return fmt.Sprintf(`<img src="%v" alt="%v">`, matches[2], matches[1])
			},
		},
		{
			regexp.MustCompile(`^([^!]*)\[([^\]]*)\]\(([^)]+)\)(.*)$`),
			func(matches []string) string {
				return fmt.Sprintf(`%v<a href="%v" target="_blank" rel="noopener noreferrer">%v</a>%v`, matches[1], matches[3], matches[2], matches[4])
			},
		},
		{
			regexp.MustCompile("^(.*)`([^`]*)`(.*)$"),
			func(matches []string) string {
				return fmt.Sprintf(`%v<code>%v</code>%v`, matches[1], matches[2], matches[3])
			},
		},
	}
)

func parsePlainSlide(content string) (output string) {
	for i := range plainSlideReplacements {
		output = plainSlideReplacements[i].regex.ReplaceAllStringFunc(content, func(match string) string {
			submatches := plainSlideReplacements[i].regex.FindStringSubmatch(match)
			return plainSlideReplacements[i].parseResult(submatches)
		})
		if output != content {
			return output
		}
	}
	return content
}

type server struct {
	commandsFile   string
	preCommands    []string
	slides         []slide
	commandManager *commandManager
	logger         *slog.Logger
}

func (s *server) GetPreCommands() []string {
	return s.preCommands
}

func (s *server) GetSlide(idx int) slide {
	return s.slides[idx]
}

func (s *server) GetSlideCount() int {
	return len(s.slides)
}

func (s *server) ParsePreCommands(contents []string) (i int, err error) {
	for i < len(contents) && !hrRegex.MatchString(contents[i]) {
		s.preCommands = append(s.preCommands, contents[i])
		i++
	}

	if i == len(contents) { // no preCommands
		s.preCommands = []string{}
		i = 0
		return
	}

	i++ // skip the separator
	return
}

func (s *server) ParseSlide(content string, t slideType) {
	s.slides = append(s.slides, slide{
		id:        len(s.slides),
		content:   content,
		slideType: t,
	})
}

func (s *server) Initialise(ctx context.Context) (err error) {
	for _, cmd := range s.GetPreCommands() {
		if strings.HasPrefix("#", cmd) {
			continue
		}

		s.logger.Info("running pre-command", "command", cmd)

		switch {
		case varCommandPattern.MatchString(cmd):
			if err := setCommandToVar(cmd); err != nil {
				err = fmt.Errorf("could not set command to var '%v': %v", cmd, err.Error())
				return err
			}
		case varPattern.MatchString(cmd):
			if err := setVar(cmd); err != nil {
				err = fmt.Errorf("could set variable '%v': %v", cmd, err.Error())
				return err
			}
		default:
			if err := runCommand(cmd); err != nil {
				err = fmt.Errorf("could not execute pre-command '%v': %v", cmd, err.Error())
				return err
			}
		}
	}

	return
}

func (s *server) ParseSlides(contents []string, startIdx int) (err error) {
	insideCodeBlock := false
	var currentCommand strings.Builder

	contents = contents[startIdx:]

	for i := range contents {
		line := contents[i]
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "```") {
			if insideCodeBlock {
				currentCommand.WriteString("\n</code></pre>")
				s.ParseSlide(currentCommand.String(), slideTypeCodeblock)
				insideCodeBlock = false
				currentCommand.Reset()
			} else {
				currentCommand.WriteString(fmt.Sprintf("<pre class='language-%v'><code>", strings.TrimPrefix(trimmed, "```")))
				insideCodeBlock = true
			}
			continue
		}

		if insideCodeBlock {
			if currentCommand.Len() > 0 {
				currentCommand.WriteString("\n")
			}
			currentCommand.WriteString(line)
			continue
		}

		switch {
		case trimmed == "":
		case strings.HasPrefix(trimmed, "$ "):
			s.ParseSlide(strings.TrimPrefix(trimmed, "$ "), slideTypeCommand)
		default:
			s.ParseSlide(parsePlainSlide(trimmed), slideTypePlain)
		}
	}

	return
}

func (s *server) SplitContent(commandsFile string) (slideContent []string, err error) {
	contents, err := os.ReadFile(commandsFile)
	if err != nil {
		return
	}

	slideContent = strings.Split(regexp.MustCompile(`\\\s*\n`).ReplaceAllString(string(contents), ""), "\n")
	slideContent = slices.DeleteFunc(slideContent, func(s string) bool { return s == "" })
	return
}

func NewServer(logger *slog.Logger, commandsFile string) (s IPresentationServer, err error) {
	s = &server{
		logger:         logger,
		commandsFile:   commandsFile,
		commandManager: newCommandManager(logger),
	}

	slideContent, err := s.SplitContent(commandsFile)
	if err != nil {
		return
	}

	startIdx, err := s.ParsePreCommands(slideContent)
	if err != nil {
		err = errors.New("could not parse pre-commands")
		return
	}

	err = s.ParseSlides(slideContent, startIdx)
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

	return http.ListenAndServe("localhost:8080", mux)
}
