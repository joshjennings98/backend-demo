package server

import (
	"fmt"

	"github.com/maragudk/gomponents"
	hx "github.com/maragudk/gomponents-htmx"
	"github.com/maragudk/gomponents/html"
)

func gomponentsIfElse(condition bool, ifBranch, elseBranch gomponents.Node) gomponents.Node {
	if condition {
		return ifBranch
	}
	return elseBranch
}

func (m *DemoManager) cleanedCommandGomponent() gomponents.Node {
	return html.P(gomponents.Raw(m.cleanedCommand()))
}

func (m *DemoManager) indexHTML(commands []string, idx int32, isCmd, isCmdRunning bool) gomponents.Node {
	return html.HTML(
		html.Head(
			html.TitleEl(gomponents.Text("Backend Demo Tool")),
			html.Script(html.Src("static/main.js")),
			html.Script(html.Src("https://unpkg.com/htmx.org")),
			html.Script(html.Src("https://cdn.jsdelivr.net/npm/xterm/lib/xterm.js")),
			html.Link(html.Rel("stylesheet"), html.Href("static/main.css")),
			html.Link(html.Rel("stylesheet"), html.Href("https://cdn.jsdelivr.net/npm/xterm/css/xterm.css")),
		),
		html.Body(
			m.contentDiv(commands, idx, isCmd, isCmdRunning),
		),
	)
}

func (m *DemoManager) contentDiv(commands []string, idx int32, isCmd, isCmdRunning bool) gomponents.Node {
	return html.Div(
		html.ID("command"),
		html.Div(
			html.ID("controls"),
			slideSelect(commands, idx),
			html.FormEl(
				html.Class("control"),
				hx.Delete(pageEndpoint),
				hx.Swap("outerHTML"),
				hx.Target("#command"),
				hx.Trigger("click, keyup[key=='ArrowLeft'] from:body"),
				html.Button(gomponents.Text("prev")),
			),
			html.FormEl(
				html.Class("control"),
				hx.Post(pageEndpoint),
				hx.Swap("outerHTML"),
				hx.Target("#command"),
				hx.Trigger("click, keyup[key=='ArrowRight'] from:body"),
				html.Button(gomponents.Text("next")),
			),
		),
		html.Div(
			html.Div(
				gomponentsIfElse(
					isCmd,
					html.Class("command-string"),
					html.Class("text-string"),
				),
				m.cleanedCommandGomponent(),
			),
		),
		html.Div(
			gomponents.If(
				!isCmd,
				gomponents.Attr("hidden", "true"),
			),
			html.Div(
				html.ID("terminal"),
				hx.Preserve("true"),
			),
			runningButton(isCmdRunning),
		),
	)
}

func slideSelect(commands []string, selected int32) gomponents.Node {
	var options []gomponents.Node
	for i := int32(0); i < int32(len(commands)); i++ {
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
		hx.Put(pageEndpoint),
		hx.Trigger("change"),
		hx.Target("#command"),
		html.Name("slideIndex"),
		gomponents.Group(options),
	)
}

func runningButton(isCmdRunning bool) gomponents.Node {
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

func stopButton() gomponents.Node {
	return html.Div(
		html.ID("stop-button"),
		html.FormEl(
			hx.Delete("/execute"),
			hx.Target("#stop-button"),
			html.Button(gomponents.Text("stop")),
		),
		html.FormEl(
			hx.Delete("/execute"),
			hx.Target("#stop-button"),
			hx.Trigger("keyup[key==' '] from:body"),
			gomponents.Attr("hidden", "true"),
		),
	)
}

func executeButton() gomponents.Node {
	return html.Div(
		html.ID("execute-button"),
		html.FormEl(
			hx.Post("/execute"),
			hx.Target("#execute-button"),
			html.Button(gomponents.Text("execute")),
		),
		html.FormEl(
			hx.Post("/execute"),
			hx.Target("#execute-button"),
			hx.Trigger("keyup[key==' '] from:body"),
			gomponents.Attr("hidden", "true"),
		),
	)
}
