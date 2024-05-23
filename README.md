# Backend Demo Presentation Tool

Present backend demos in a powerpoint-like format, written in Go using HTMX and websockets. 

## Background

Frontend developers have a much easier time of it when it comes to demonstrating their work. It is easy to see what they've done and they can make it look as interesting as they want. In comparison, trying to demo a backend API or tool is rarely interesting to watch. It usually involves a lot of text, ends up with the presenter pulling out a terminal to show commands on the fly, searching their command history and hoping they do everything in the right order. All this to get some JSON or similarly uninteresting output.

This project is an attempt to try and improve that situation and make backend demos simpler and more engaging. To do this, an application was created that reads a simple text file containing information to present and generates a web server with Go and [HTMX](https://htmx.org/). The tool is able to create a simple presentation that is shown in the browser and can be clicked through like a PowerPoint presentation with the content being served using HTMX. Code blocks can be shown and will be automatically highlighted with [highlight.js](https://highlightjs.org/). Commands can be run on the host machine and streamed to a terminal in the browser thanks to websockets and [xterm.js](http://xtermjs.org/).

This allows you to easily demonstrate commands without having to open a terminal. It also means you can plan the exact order of the commands and not worry about mixing them up. A side effect of running the presentation in the browser is that you can easily have links that open up in new tabs which can be useful to show the result of any executed commands, this is useful as you can just share the web browser during a meeting instead of having to switch between PowerPoint, a terminal, and a web browser.

This application lends itself to a presentation method similar to the [Takahashi method](https://en.wikipedia.org/wiki/Takahashi_method) where you have a concise slides with very little text. The created presentations are more complex than the Taskahashi method but it is still purposefully designed to keep things simple by only having one type of content per slide, be that text, a command output, or something else.

The result is that with this application backend demos should be much easier to create and run whilst also being slightly more engaging.
Features

* A presentation is just a simple text file.
* A slide can contain: text, an executed command, a code block, or an image. The syntax uses markdown.
* Executed commands will stream output to be shown in the slide thanks to [xterm.js](http://xtermjs.org/).
* The command that is executed is shown above the command output.
* Shell commands can be executed before the presentation runs allowing for setup before a presentation.
* Code blocks are automatically highlighted using [highlight.js](https://highlightjs.org/).
* Embed HTML meaning you can include videos or iframes etc.
* Left click or right arrow to go forwards, right click or left arrow to go backwards.
* Easy to use and share via screen sharing.
* Styling done via CSS meaning it can be easily reconfigured.
* Jump to any slide via a drop down menu.

## Creating a presentation

Write a command file `commands.txt` like this:

```md
# setup environment here (comments will be ignored)
VAR=hello # You can set variables
VAR2=$(echo world) # you can also set them to the output of a command using $()
echo hello world # or you can just run commands
---

Text only slides are just lines of plain text

![Specify images like in markdown](https://upload.wikimedia.org/wikipedia/en/7/73/Hyperion_cover.jpg)

$ echo 'run a command in the same shell as the current presentation process by prefixing the line with $'

$ echo 'long running commands are also supported' && sleep 5 && echo 'thanks to xterm.js' && sleep 5 && echo 'they will not block the presentation'

$ echo $VAR # This variable was set up earlier

```go
func main() {
    fmt.Println("Use fenced code blocks with markdown syntax to get")
    fmt.Println("a block of syntax highlighted code with highlight.js")
}
`​``
```

One line per slide to keep things simple (although you can do multiple lines with backslashes or `<br>`.

Another example can be found in [commands.txt](./server/testdata/commands.txt).

## Installation

### Nix Flake

If you are on `nix` then you can do try it out with:

```sh
nix run github:joshjennings98/backend-demo --no-write-lock-file
```

Or install it [using flakes](https://nixos.wiki/wiki/Flakes).

### Manual Build

Clone the repository and `cd` into `cli`. Then run:

```sh
go build . && mv cli /bin/backend-demo
```

### Releases

Go to the [releases page](https://github.com/joshjennings98/backend-demo/releases) and install one of the releases.

This has not been tests so please raise an issue if they do not work.

## Usage

Pass it as an argument to `backend-demo`:

```
backend-demo -c commands.txt
```

Use the mouse button to go forward and back or select a slide via the dropdown menu.

Alternatively use the arrow keys for forward and back and the space bar to execute the command.
