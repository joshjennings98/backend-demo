package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
)

const (
	indexEndpoint         = "GET /presentation"
	initEndpoint          = "GET /init"
	pageEndpoint          = "GET /slides/{id}/"
	startCommandEndpoint  = "POST /commands/{id}/"
	statusCommandEndpoint = "GET /commands/{id}/"
	stopCommandEndpoint   = "DELETE /commands/{id}/"
)

var (
	hrRegex = regexp.MustCompile(`^\-+`)
)

type server struct {
	preCommands    []string
	slides         []*slide
	commandManager *commandManager
}

func (s *server) getPreCommands() []string {
	return s.preCommands
}

func (s *server) getSlide(idx int) *slide {
	return s.slides[idx]
}

func (s *server) totalSlides() int {
	return len(s.slides)
}

func (s *server) parsePreCommands(contents []string) (i int, err error) {
	for !hrRegex.MatchString(contents[i]) {
		s.preCommands = append(s.preCommands, contents[i])
		i++
	}
	return
}

func (s *server) addSlide(content string, t slideType) {
	s.slides = append(s.slides, &slide{
		id:        len(s.slides),
		content:   content,
		slideType: t,
	})
}

func (s *server) initialise(ctx context.Context) (err error) {
	for _, cmd := range s.getPreCommands() {
		log.Println("running pre-command:", cmd)
		if varCommandPattern.MatchString(cmd) {
			if err := setCommandToVar(cmd); err != nil {
				err = fmt.Errorf("could not set command to var '%v': %v", cmd, err.Error())
				return err
			}
		} else if varPattern.MatchString(cmd) {
			if err := setVar(cmd); err != nil {
				err = fmt.Errorf("could set variable '%v': %v", cmd, err.Error())
				return err
			}
		} else {
			if err := runCommand(cmd); err != nil {
				err = fmt.Errorf("could not execute pre-command '%v': %v", cmd, err.Error())
				return err
			}
		}
	}

	return
}

func (s *server) parseSlides(contents []string, startIdx int) (err error) {
	insideCodeBlock := false
	var currentCommand strings.Builder

	contents = contents[startIdx+1:]

	for i := range contents {
		line := contents[i]
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "```") {
			if insideCodeBlock {
				currentCommand.WriteString("\n```")
				s.addSlide(currentCommand.String(), slideTypeCodeblock)
				insideCodeBlock = false
				currentCommand.Reset()
			} else {
				currentCommand.WriteString(line)
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
			s.addSlide(strings.TrimPrefix(trimmed, "$ "), slideTypeCommand)
		default:
			s.addSlide(trimmed, slideTypePlain)
		}
	}

	return
}

func NewServer(presentationPath string) (s IServer, err error) {
	contents, err := os.ReadFile(presentationPath)
	if err != nil {
		return
	}

	s = &server{
		commandManager: &commandManager{
			mu: sync.Mutex{},
		},
	}

	slideContent := strings.Split(regexp.MustCompile(`\\\s*\n`).ReplaceAllString(string(contents), ""), "\n")

	startIdx, err := s.parsePreCommands(slideContent)
	if err != nil {
		err = errors.New("could not parse pre-commands")
		return
	}

	err = s.parseSlides(slideContent, startIdx)
	if err != nil {
		err = errors.New("could not parse slides")
		return
	}

	return
}

func Start(ctx context.Context, presentationPath string) error {
	m, err := NewServer(presentationPath)
	if err != nil {
		return err
	}

	err = m.initialise(ctx)
	if err != nil {
		return err
	}

	r := http.NewServeMux()

	r.HandleFunc(indexEndpoint, m.indexHandler)
	r.HandleFunc(initEndpoint, m.initHandler)
	r.HandleFunc(pageEndpoint, m.showSlideHandler)
	r.HandleFunc(startCommandEndpoint, m.startCommandHandler)
	r.HandleFunc(statusCommandEndpoint, m.statusCommandHandler)
	r.HandleFunc(stopCommandEndpoint, m.stopCommandHandler)

	r.HandleFunc("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))).ServeHTTP)

	log.Println("Server is running on http://localhost:8080/presentation")
	_ = http.ListenAndServe("localhost:8080", r)

	return nil
}
