from flask import Flask, render_template, Response, jsonify
import subprocess

from .utils import Tag, convert_to_html
from .config import Config

app = Flask(__name__)

@app.route('/')
def index():
    return render_template(
        'index.html', 
        command_tag=Tag.COMMAND.value, 
        text_tag=Tag.TEXT.value, 
        image_tag=Tag.IMAGE.value, 
        code_tag=Tag.CODE.value
    )

@app.route('/pages')
def get_pages():
    return jsonify(pages=Config.LINES)

@app.route('/command/<int:index>')
def current_command(index):
    line = Config.LINES[index]
    return '<br>'.join(line['content']) if line['type'] == Tag.COMMAND else ""

@app.route('/pages/<int:index>')
def page(index):
    line = Config.LINES[index]
    if line['type'] == Tag.COMMAND:
        def generate():
            yield '<pre style="font-size: 18px; background-color: black; color: white;">'
            with subprocess.Popen(''.join(line['content']), stdout=subprocess.PIPE, stderr=subprocess.STDOUT, bufsize=1, universal_newlines=True, shell=True) as p:
                for output_line in p.stdout:
                    yield f"{output_line.rstrip()}\n"
                    yield '<script>window.scrollTo(0,document.body.scrollHeight);</script>'
            yield '</pre>'
        return Response(generate(), mimetype='text/html')
    if line['type'] == Tag.CODE:
        content = '\n'.join(line['content']).strip()
        return f"<pre><code>{content}</code></pre>"
    if line['type'] == Tag.IMAGE:
        content = ''.join(line['content'])
        return f"<img src='{content}'>"
    else:
        text_lines = [convert_to_html(line) for line in line['content'] if line]
        return jsonify({'text_lines': text_lines})
