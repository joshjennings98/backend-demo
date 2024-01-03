package server

import (
	"fmt"
	"strings"

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
