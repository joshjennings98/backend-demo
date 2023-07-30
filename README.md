# Backend Demos Script

Present backend demos in a powerpoint-like format.


## Installation

To install go into the cloned repo and run:

```
pip install .
```

This will install the tool as `backend_demo`.

## Usage

Write a command file `commands.txt` like this:

```
[TEXT] 
This is a slide

[COMMAND] 
ping -c 4 google.com

[TEXT] 
I am another text slide
with multiple lines

[IMAGE]
https://upload.wikimedia.org/wikipedia/en/7/73/Hyperion_cover.jpg

[CODE]
print("I am a code block, I will be automatically syntax highlighted with hightlight.js")
```

Pass it as an argument to `backend_demo`:

```
backend_demo commands.txt
```

## Bugs

Fast long running commands like `ls -R /` are quite glitchy.

## To Do

* Move template stuff and css stuff to templates and static directories.
* Don't use `shell=True` but retain the functionality that provides (pipes and redirects etc.)
* Improve long running fast commands (e.g. `ls -R /`), possible by caching output and reading that.
