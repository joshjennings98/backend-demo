package server

import (
	"strings"
	"testing"

	"github.com/maragudk/gomponents/html"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNextSlide(t *testing.T) {
	assert.Equal(t, 4, nextSlide(3, 5))
	assert.Equal(t, 0, nextSlide(4, 5))
}

func TestPrevSlide(t *testing.T) {
	assert.Equal(t, 3, prevSlide(4, 5))
	assert.Equal(t, 4, prevSlide(0, 5))
}

func TestContentDiv(t *testing.T) {
	t.Run("plain slide", func(t *testing.T) {
		testSlide := slide{id: 0, content: "this is a some text", slideType: slideTypePlain}
		var actual strings.Builder
		err := contentDiv(3, 10, testSlide, false).Render(&actual)
		require.NoError(t, err)
		expected := `<div id="command"><div id="controls"><select hx-get="/slides" hx-target="#command" name="idx"><option value="0">Slide 1/10</option><option value="1">Slide 2/10</option><option value="2">Slide 3/10</option><option value="3" selected>Slide 4/10</option><option value="4">Slide 5/10</option><option value="5">Slide 6/10</option><option value="6">Slide 7/10</option><option value="7">Slide 8/10</option><option value="8">Slide 9/10</option><option value="9">Slide 10/10</option></select><form class="control" hx-get="/slides/2" hx-swap="outerHTML" hx-target="#command" hx-trigger="click, keyup[key==&#39;ArrowLeft&#39;] from:body"><button>prev</button></form><form class="control" hx-get="/slides/4" hx-swap="outerHTML" hx-target="#command" hx-trigger="click, keyup[key==&#39;ArrowRight&#39;] from:body"><button>next</button></form></div><div><div class="text-string"><p>this is a some text</p></div></div><div hidden="true"><div id="terminal" hx-preserve="true"></div><div><div id="execute-button"><form hx-post="/commands/0/start" hx-target="#execute-button"><button>execute</button></form><form hx-post="/commands/0/start" hx-target="#execute-button" hx-trigger="keyup[key==&#39; &#39;] from:body" hidden="true"></form></div></div></div></div>`
		assert.Equal(t, expected, actual.String())
	})

	t.Run("code slide", func(t *testing.T) {
		testSlide := slide{id: 0, content: "<pre><code>this is some code</code></pre>", slideType: slideTypeCodeblock}
		var actual strings.Builder
		err := contentDiv(3, 10, testSlide, false).Render(&actual)
		require.NoError(t, err)
		expected := `<div id="command"><div id="controls"><select hx-get="/slides" hx-target="#command" name="idx"><option value="0">Slide 1/10</option><option value="1">Slide 2/10</option><option value="2">Slide 3/10</option><option value="3" selected>Slide 4/10</option><option value="4">Slide 5/10</option><option value="5">Slide 6/10</option><option value="6">Slide 7/10</option><option value="7">Slide 8/10</option><option value="8">Slide 9/10</option><option value="9">Slide 10/10</option></select><form class="control" hx-get="/slides/2" hx-swap="outerHTML" hx-target="#command" hx-trigger="click, keyup[key==&#39;ArrowLeft&#39;] from:body"><button>prev</button></form><form class="control" hx-get="/slides/4" hx-swap="outerHTML" hx-target="#command" hx-trigger="click, keyup[key==&#39;ArrowRight&#39;] from:body"><button>next</button></form></div><div><div class="text-string"><p><pre><code>this is some code</code></pre></p></div></div><div hidden="true"><div id="terminal" hx-preserve="true"></div><div><div id="execute-button"><form hx-post="/commands/0/start" hx-target="#execute-button"><button>execute</button></form><form hx-post="/commands/0/start" hx-target="#execute-button" hx-trigger="keyup[key==&#39; &#39;] from:body" hidden="true"></form></div></div></div></div>`
		assert.Equal(t, expected, actual.String())
	})

	t.Run("comand slide", func(t *testing.T) {
		testSlide := slide{id: 0, content: "this is a command", slideType: slideTypeCommand}
		var actual strings.Builder
		err := contentDiv(3, 10, testSlide, false).Render(&actual)
		require.NoError(t, err)
		expected := `<div id="command"><div id="controls"><select hx-get="/slides" hx-target="#command" name="idx"><option value="0">Slide 1/10</option><option value="1">Slide 2/10</option><option value="2">Slide 3/10</option><option value="3" selected>Slide 4/10</option><option value="4">Slide 5/10</option><option value="5">Slide 6/10</option><option value="6">Slide 7/10</option><option value="7">Slide 8/10</option><option value="8">Slide 9/10</option><option value="9">Slide 10/10</option></select><form class="control" hx-get="/slides/2" hx-swap="outerHTML" hx-target="#command" hx-trigger="click, keyup[key==&#39;ArrowLeft&#39;] from:body"><button>prev</button></form><form class="control" hx-get="/slides/4" hx-swap="outerHTML" hx-target="#command" hx-trigger="click, keyup[key==&#39;ArrowRight&#39;] from:body"><button>next</button></form></div><div><div class="command-string"><p>this is a command</p></div></div><div><div id="terminal" hx-preserve="true"></div><div><div id="execute-button"><form hx-post="/commands/0/start" hx-target="#execute-button"><button>execute</button></form><form hx-post="/commands/0/start" hx-target="#execute-button" hx-trigger="keyup[key==&#39; &#39;] from:body" hidden="true"></form></div></div></div></div>`
		assert.Equal(t, expected, actual.String())
	})
}

func TestIndex(t *testing.T) {
	content := html.Div()
	var actual strings.Builder
	err := indexHTML(content).Render(&actual)
	require.NoError(t, err)
	expected := `<html><head><title>Backend Demo Tool</title><meta name="viewport" content="width=device-width, initial-scale=1.0"><script src="static/main.js"></script><script src="static/highlight.js"></script><script src="static/htmx.js"></script><script src="static/xterm.js"></script><link rel="stylesheet" href="static/main.css"><link rel="stylesheet" href="static/xterm.css"><link rel="stylesheet" href="static/highlight.css"></head><body><div></div></body></html>`
	assert.Equal(t, expected, actual.String())
}
