# Backend Demos Script

Present backend demos in a powerpoint-like format.

## Usage

Write a command file like this:

```
TXT This is a presentation
CMD ping -c 4 google.com
TXT I am a slide<br>with multiple lines
```

Pass it as an argument to `present.py`:

```
python3 present.py commands.txt
```

## Bugs

Fast long running commands like `ls -R /` are quite glitchy.

## To Do

* Add support for `IMG` prefix for showing a single image on a slide.
* Add support for `CDE` prefix for showing a block of code on a slide.
* Make it so that `<br>` isn't needed for multi-line stuff.
* Move template stuff and css stuff to templates and static directories.
* Package up the tool so it is easy to run, [see this page](http://blog.ablepear.com/2012/10/bundling-python-files-into-stand-alone.html).
* Add release workflow and releases.
