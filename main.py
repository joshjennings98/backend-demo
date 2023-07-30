import argparse
import webbrowser

from app import app
from utils import read_file, parse_commands
from config import Config

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description='Process a commands file.')
    parser.add_argument('file_path', type=str, help='The path to the commands file')
    args = parser.parse_args()

    Config.LINES = parse_commands(read_file(args.file_path))

    webbrowser.open_new_tab(f"http://localhost:{Config.PORT}")
    app.run(port=Config.PORT, debug=False)
