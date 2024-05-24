package server

import (
	"fmt"
	"strings"

	"github.com/maragudk/gomponents"
	hx "github.com/maragudk/gomponents-htmx"
	"github.com/maragudk/gomponents/html"

	"github.com/joshjennings98/backend-demo/server/v2/types"
)

func gomponentsIfElse(condition bool, ifBranch, elseBranch gomponents.Node) gomponents.Node {
	if condition {
		return ifBranch
	}
	return elseBranch
}

func cleanedCommandGomponent(command string, slideType types.SlideType) gomponents.Node {
	if slideType != types.SlideTypeCommand {
		return gomponents.Raw(strings.TrimSpace(command))
	}
	return html.P(gomponents.Raw(strings.TrimSpace(command)))
}

func indexHTML(content gomponents.Node) gomponents.Node {
	return html.HTML(
		html.Head(
			html.TitleEl(gomponents.Text("Backend Demo Tool")),
			html.Meta(html.Name("viewport"), html.Content("width=device-width, initial-scale=1.0")),
			html.Script(html.Src("static/main.js")),
			html.Script(html.Src("static/highlight.js")),
			html.Script(html.Src("static/htmx.js")),
			html.Script(html.Src("static/xterm.js")),
			html.Link(html.Rel("stylesheet"), html.Href("static/main.css")),
			html.Link(html.Rel("stylesheet"), html.Href("static/xterm.css")),
			html.Link(html.Rel("stylesheet"), html.Href("static/highlight.css")),
		),
		html.Body(
			content,
		),
	)
}

func nextSlide(i, total int) int {
	return (i + 1) % total
}

func prevSlide(i, total int) int {
	return ((i-1)%total + total) % total
}

func contentDiv(slideIdx, totalSlides int, command types.Slide, isCmdRunning bool) gomponents.Node {
	return html.Div(
		html.ID("command"),
		html.Div(
			html.ID("controls"),
			slideSelect(slideIdx, totalSlides),
			html.FormEl(
				html.Class("control"),
				hx.Get(fmt.Sprintf("/slides/%v", prevSlide(slideIdx, totalSlides))),
				hx.Swap("outerHTML"),
				hx.Target("#command"),
				hx.Trigger("click, keyup[key=='ArrowLeft'] from:body"),
				html.Button(gomponents.Text("prev")),
			),
			html.FormEl(
				html.Class("control"),
				hx.Get(fmt.Sprintf("/slides/%v", nextSlide(slideIdx, totalSlides))),
				hx.Swap("outerHTML"),
				hx.Target("#command"),
				hx.Trigger("click, keyup[key=='ArrowRight'] from:body"),
				html.Button(gomponents.Text("next")),
			),
		),
		html.Div(
			html.Div(
				gomponentsIfElse(
					command.SlideType == types.SlideTypeCommand,
					html.Class("command-string"),
					html.Class("text-string"),
				),
				cleanedCommandGomponent(command.Content, command.SlideType),
			),
		),
		html.Div(
			gomponents.If(
				command.SlideType != types.SlideTypeCommand,
				gomponents.Attr("hidden", "true"),
			),
			html.Div(
				html.ID("terminal"),
				hx.Preserve("true"),
			),
			runningButton(command.ID, isCmdRunning),
		),
	)
}

func slideSelect(slideIdx, totalSlides int) gomponents.Node {
	var options []gomponents.Node
	for i := 0; i < totalSlides; i++ {
		options = append(options, html.Option(
			gomponents.Group([]gomponents.Node{
				html.Value(fmt.Sprint(i)),
				gomponents.If(
					i == slideIdx,
					html.Selected(),
				),
			}),
			gomponents.Text(fmt.Sprintf("Slide %d/%d", i+1, totalSlides)),
		))
	}

	return html.Select(
		hx.Get("/slides"),
		hx.Target("#command"),
		html.Name("idx"),
		gomponents.Group(options),
	)
}

func runningButton(idx int, isCmdRunning bool) gomponents.Node {
	if isCmdRunning {
		return html.Div(
			hx.Get(fmt.Sprintf("/commands/%v/status", idx)),
			hx.Trigger("every 100ms"),
			hx.Target("#execute-button"),
			hx.Swap("outerHTML"),
			stopButton(idx),
		)
	}

	return html.Div(
		executeButton(idx),
	)
}

func stopButton(idx int) gomponents.Node {
	return html.Div(
		html.ID("stop-button"),
		html.FormEl(
			hx.Post(fmt.Sprintf("/commands/%v/stop", idx)),
			hx.Target("#stop-button"),
			html.Button(gomponents.Text("stop")),
		),
		html.FormEl(
			hx.Post(fmt.Sprintf("/commands/%v/stop", idx)),
			hx.Target("#stop-button"),
			hx.Trigger("keyup[key==' '] from:body"),
			gomponents.Attr("hidden", "true"),
		),
	)
}

func executeButton(idx int) gomponents.Node {
	return html.Div(
		html.ID("execute-button"),
		html.FormEl(
			hx.Post(fmt.Sprintf("/commands/%v/start", idx)),
			hx.Target("#execute-button"),
			html.Button(gomponents.Text("execute")),
		),
		html.FormEl(
			hx.Post(fmt.Sprintf("/commands/%v/start", idx)),
			hx.Target("#execute-button"),
			hx.Trigger("keyup[key==' '] from:body"),
			gomponents.Attr("hidden", "true"),
		),
	)
}
