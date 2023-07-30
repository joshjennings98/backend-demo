# Backend Demos Script

Present backend demos in a powerpoint-like format.

## Usage

Write a command file like this:

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

Pass it as an argument to `main.py`:

```
python3 main.py commands.txt
```

## Bugs

Fast long running commands like `ls -R /` are quite glitchy.

## To Do

* Move template stuff and css stuff to templates and static directories.
* Package up the tool so it is easy to run, [see this page](http://blog.ablepear.com/2012/10/bundling-python-files-into-stand-alone.html).
* Add release workflow and releases.
* Improve long running fast commands (e.g. `ls -R /`), possible by caching output and reading that.
