from flask import Flask, render_template_string, Response, jsonify
import argparse, re, subprocess, webbrowser
from enum import Enum
 
class Tag(str, Enum):
    COMMAND = "[COMMAND]"
    TEXT    = "[TEXT]"
    CODE    = "[CODE]"
    IMAGE   = "[IMAGE]"

    @classmethod
    def is_tag(cls, key):
        return key in [member.value for member in Tag]
    
app = Flask(__name__)

def read_file(file_path):
    try:
        with open(file_path, 'r') as f:
            return f.readlines()
    except Exception as e:
        print(f"Error while reading file: {e}")
        return []

def parse_commands(lines):
    commands = []
    current_command = None
    current_content = []
    for line in lines:
        line = line.strip()

        if Tag.is_tag(line):
            if current_command:
                commands.append({'type': current_command, 'content': current_content})
            current_command = line
            current_content = []
        else:
            current_content.append(line)

    commands.append({'type': current_command, 'content': current_content})

    return commands

def convert_to_html(text):
    text = re.sub(r"`(.*?)`", r"<code>\1</code>", text) # Convert backticks to <code></code>
    text = re.sub(r"\[(.*?)\]\((.*?)\)", r'<a href="\2">\1</a>', text) # Convert markdown links to html links
    return text

@app.route('/')
def index():
    return render_template_string("""
<!DOCTYPE html>
<html>
    <head>
        <title>Command Slides</title>
        <base target="_blank">
        <link rel="stylesheet" href="//cdnjs.cloudflare.com/ajax/libs/highlight.js/11.3.1/styles/default.min.css">
        <style>
            html, body {
                width: 99%;
                height: 95%;
                padding: 0;
            }

            #app {
                width: 80%;
                height: 90%;
                margin: 50px auto;
                display: flex;
                justify-content: center;
                align-items: center;
                flex-direction: column;
            }

            #command-display {
                background-color: black;
                height: 100%;
                display: none;
                width: 100%;
            }
    
            pre {
                font-family: Monospace;
                padding: 10px;
                border: black 10px;
                font-size: 24px;
            }

            #text-output {
                box-sizing: border-box;
                font-size: 36px;
                font-family: Arial;
                user-select: none;
                text-align: center;
                display: none;
                width: 100%;
            }

            #current-command {
                display: none;
                font-family: Monospace;
                text-align: center;
                padding: 10px;
                background-color: black;
                color: white;
                font-size: 24px;
                margin-bottom: 20px;
            }

            code {
                font-family: Monospace;
                background-color: #f0f0f0;
                padding: 3px;
            }
        </style>
    </head>
    <body>
        <div id="controls">
            <select id="command-select"></select>
        </div>
        <main id="app">  
            <div id="current-command"></div>
            <iframe id="command-display" frameBorder="0"></iframe>
            <div id="text-output"></div>
        </main>
        <script>
            const iframe = document.getElementById('command-display');
            const outputDiv = document.getElementById('text-output');
            const commandSelect = document.getElementById('command-select');
            const currentCommand = document.getElementById('current-command');
            let commands = [];
            let textLines = [];
            let lineIndex = 0;
            let displayedLines = [];
            let index = 0;
            let navigationDirection = 'forward';

            async function fetchCommands() {
                const response = await fetch('/commands');
                const data = await response.json();
                commands = data.commands;

                // Generate select options
                commands.forEach((command, i) => {
                    const option = document.createElement('option');
                    option.value = i;
                    option.textContent = `Slide ${i + 1}/${commands.length}`;
                    commandSelect.appendChild(option);
                });

                loadCommand();
            }

            async function loadCommand() {
                outputDiv.style.display = 'none';
                iframe.style.display = 'none';
                currentCommand.style.display = 'none';

                if (commands[index].type === '"""+Tag.COMMAND+"""') {
                    iframe.src = `/command/${index}`;
                    iframe.style.display = 'block';
                    currentCommand.style.display = 'block';
                    textLines = [];
                    displayedLines = [];
                    lineIndex = 0;
                } else if (commands[index].type === '"""+Tag.CODE+"""' || commands[index].type === '"""+Tag.IMAGE+"""') {
                    const response = await fetch(`/command/${index}`);
                    const data = await response.text();
                    outputDiv.innerHTML = data;
                    if (commands[index].type === '"""+Tag.CODE+"""') {
                        hljs.highlightAll();
                    }
                    outputDiv.style.display = 'block';
                } else {
                    const response = await fetch(`/command/${index}`);
                    const data = await response.json();
                    textLines = data.text_lines;
                    if (navigationDirection === 'backward') {
                        // If we're navigating backward, display all lines
                        lineIndex = textLines.length - 1;
                        displayedLines = [...textLines];
                    } else {
                        // If we're navigating forward, display the first line
                        lineIndex = 0;
                        displayedLines = [textLines[lineIndex]];
                    }
                    outputDiv.innerHTML = displayedLines.join('<br><br><br>');
                    outputDiv.style.height = `${3 * textLines.length}em`;
                    outputDiv.style.display = 'block';
                }

                commandSelect.value = index;
                currentCommand.textContent = "";

                // Fetch and display current command
                const commandResponse = await fetch(`/current_command/${index}`);
                const commandData = await commandResponse.text();
                currentCommand.textContent = commandData;
            }

            commandSelect.addEventListener('change', () => {
                index = Number(commandSelect.value);
                loadCommand();
            });

            document.body.addEventListener('mousedown', (event) => {
                const controlsDiv = document.getElementById('controls');
                let target = event.target;

                while (target != null) {
                    if (target === controlsDiv || (target.tagName != null && target.tagName.toLowerCase() === "a")) {
                        return;
                    }
                    target = target.parentNode;
                }

                if (event.button === 2) { // Right-click
                    handleRightClick();
                } else if (event.button === 0) { // Left-click
                    handleLeftClick();
                }
            });

            // Prevent the context menu from showing on right-click
            document.body.addEventListener('contextmenu', (event) => {
                event.preventDefault();
            });

            // Fetch commands on load
            fetchCommands();

            // Right click to go backwards
            function handleRightClick() {
                if (commands[index].type === '"""+Tag.TEXT+"""' && displayedLines.length > 1) {
                    displayedLines.pop();
                    lineIndex--;
                    outputDiv.innerHTML = displayedLines.join('<br><br><br>');
                } else if (index > 0) {
                    index--;
                    navigationDirection = 'backward';
                    loadCommand();
                }
            }

            // Left click to go forwards
            function handleLeftClick() {
                if (commands[index].type === '"""+Tag.TEXT+"""' && lineIndex < textLines.length - 1) {
                    lineIndex++;
                    displayedLines.push(textLines[lineIndex]);
                    outputDiv.innerHTML = displayedLines.join('<br><br><br>');
                } else if (index < commands.length - 1) {
                    index++;
                    navigationDirection = 'forward';
                    loadCommand();
                }
            }
        </script>
        <script src="//cdnjs.cloudflare.com/ajax/libs/highlight.js/11.3.1/highlight.min.js"></script>
    </body>
</html>
    """)

@app.route('/commands')
def get_commands():
    return jsonify(commands=lines)

@app.route('/current_command/<int:index>')
def current_command(index):
    line = lines[index]
    if line['type'] == Tag.COMMAND:
        return line['content'][0]
    else:
        return ""

@app.route('/command/<int:index>')
def command(index):
    line = lines[index]
    if line['type'] == Tag.COMMAND:
        def generate():
            yield '<pre style="font-size: 18px; background-color: black; color: white;">'
            with subprocess.Popen(line['content'][0].split(), stdout=subprocess.PIPE, stderr=subprocess.STDOUT, bufsize=1, universal_newlines=True) as p:
                for output_line in p.stdout:
                    yield output_line.rstrip() + '\n'
                    yield '<script>window.scrollTo(0,document.body.scrollHeight);</script>'
            yield '</pre>'
        return Response(generate(), mimetype='text/html')
    if line['type'] == Tag.CODE:
        return "<pre><code>" + '\n'.join(line['content']) + "</code></pre>"
    if line['type'] == Tag.IMAGE:
        return "<img src='" + ''.join(line['content']) + "'>"
    else:
        text_lines = [convert_to_html(line) for line in line['content'] if line != ""]
        return jsonify({'text_lines': text_lines})

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description='Process a commands file.')
    parser.add_argument('file_path', type=str, help='The path to the commands file')
    args = parser.parse_args()

    lines = parse_commands(read_file(args.file_path))

    port = 1337
    webbrowser.open_new_tab(f"http://localhost:{port}")
    app.run(port=port, debug=False)
    