import re

from enum import Enum

class Tag(str, Enum):
    COMMAND = "[COMMAND]"
    TEXT    = "[TEXT]"
    CODE    = "[CODE]"
    IMAGE   = "[IMAGE]"

    @classmethod
    def is_tag(cls, key):
        return key in [member.value for member in Tag]

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
    
    if len(lines) == 0:
        return commands

    for line in lines:
        stripped = line.strip()
        if Tag.is_tag(stripped):
            if current_command:
                commands.append({'type': current_command, 'content': current_content})
            current_command = stripped
            current_content = []
        elif stripped or current_command == Tag.CODE:
            current_content.append(line)

    commands.append({'type': current_command, 'content': current_content})

    return commands

def convert_to_html(text):
    text = re.sub(r"`(.*?)`", r"<code>\1</code>", text) 
    text = re.sub(r"\[(.*?)\]\((.*?)\)", r'<a href="\2">\1</a>', text) 
    return text

