import unittest
from unittest.mock import patch, mock_open
from backend_demo.utils import Tag, read_file, parse_commands, convert_to_html

class TestUtils(unittest.TestCase):

    def test_is_tag_valid(self):
        for tag in Tag:
            with self.subTest(tag=tag):
                self.assertTrue(Tag.is_tag(tag.value))

    def test_is_tag_invalid(self):
        self.assertFalse(Tag.is_tag("[INVALID_TAG]"))
        self.assertFalse(Tag.is_tag(""))
        self.assertFalse(Tag.is_tag("COMMAND"))

    @patch('builtins.open', new_callable=mock_open, read_data="line 1\nline 2\n")
    def test_read_file_success(self, mock_file):
        result = read_file('dummy.txt')
        self.assertEqual(result, ['line 1\n', 'line 2\n'])

    @patch('builtins.print')
    def test_read_file_nonexistent(self, mock_print):
        result = read_file('nonexistent.txt')
        self.assertEqual(result, [])
        mock_print.assert_called_with("Error while reading file: [Errno 2] No such file or directory: 'nonexistent.txt'")

    @patch('builtins.open', new_callable=mock_open, read_data="")
    def test_read_file_empty(self, mock_file):
        result = read_file('empty.txt')
        self.assertEqual(result, [])

    def test_parse_commands_normal(self):
        lines = ["[COMMAND]", "echo 1", "[TEXT]", "Some text", "[CODE]", "print('Hello')"]
        result = parse_commands(lines)
        expected = [
            {'type': '[COMMAND]', 'content': ['echo 1']},
            {'type': '[TEXT]', 'content': ['Some text']},
            {'type': '[CODE]', 'content': ["print('Hello')"]}
        ]
        self.assertEqual(result, expected)

    def test_parse_commands_empty(self):
        self.assertEqual(parse_commands([]), [])

    def test_parse_commands_no_tags(self):
        lines = ["echo 1", "Some text"]
        self.assertEqual(parse_commands(lines), [{'type': None, 'content': ["echo 1", "Some text"]}])

    def test_convert_to_html_normal(self):
        text = "This is `code` and [a link](http://example.com)"
        result = convert_to_html(text)
        expected = 'This is <code>code</code> and <a href="http://example.com">a link</a>'
        self.assertEqual(result, expected)

    def test_convert_to_html_empty(self):
        self.assertEqual(convert_to_html(""), "")

    def test_convert_to_html_no_markdown(self):
        text = "Regular text without markdown"
        self.assertEqual(convert_to_html(text), text)

