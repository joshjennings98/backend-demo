# Backend Demo Presentation Tool

Present backend demos in a powerpoint-like format.

## Usage

Write a command file `commands.txt` like this:

```
# setup environment here
VAR=8
---

Text only slides are just lines of plain text

![Specify images like in markdown](https://upload.wikimedia.org/wikipedia/en/7/73/Hyperion_cover.jpg)

$ echo 'run a command in the same shell as the current presentation process by prefixing the line with $'

$ echo 'long running commands are also supported' && sleep 5 && echo 'thanks to xterm.js' && sleep 5 && echo 'they will not block the presentation'

$ echo $VAR # This variable was set up earlier

```go
func main() {
    fmt.Println("Use fenced code blocks to get a block of syntax highlighted code with highlight.js")
}
` ``
```

One line per slide to keep things simple (although you can do multiple lines with backslashes or `<br>.

Pass it as an argument to `backend_demo`:

```
backend_demo -c commands.txt
```

Use the mouse button to go forward and back or select a slide via the dropdown menu.

Alternatively use the arrow keys for forward and back and the space bar to execute the command.

## Todo

* Add tests
* General tidy up and restructure managers
* Make it easier to serve static files
* Bundle required javascript and CSS instead of using CDNs
* Probably a lot more too...
