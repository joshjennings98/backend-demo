import unittest, subprocess
from flask_testing import TestCase
from backend_demo.app import app
from backend_demo.config import Config
from backend_demo.utils import Tag
from unittest.mock import patch
from flask import jsonify

class TestApp(TestCase):

    def create_app(self):
        app.config['TESTING'] = True
        return app
    
    def setUp(self):
        Config.LINES = [
            { "type": Tag.CODE, "content": [] },
            { "type": Tag.CODE, "content": ["echo 1"] },
            { "type": Tag.CODE, "content": ["echo 2", "echo 3"] },
            { "type": Tag.TEXT, "content": [] },
            { "type": Tag.TEXT, "content": ["echo 2"] },
            { "type": Tag.TEXT, "content": ["echo 2", "echo 3"] },
            { "type": Tag.IMAGE, "content": [] },
            { "type": Tag.IMAGE, "content": ["test"] },
            { "type": Tag.IMAGE, "content": ["test1", "test2"] },
            { "type": Tag.COMMAND, "content": ["echo 1"] },
            { "type": Tag.COMMAND, "content": ["echo 2", "echo 3"] },
        ]

    def test_index(self):
        response = self.client.get('/')
        self.assertEqual(response.status_code, 200)

    def test_get_pages(self):
        response = self.client.get('/pages')
        self.assertEqual(response.status_code, 200)
        self.assertEqual(response.json, {'pages': Config.LINES})

    def test_page_route(self):
        for i, test in enumerate(Config.LINES):
            response = self.client.get(f"/pages/{i}")
            self.assertEqual(response.status_code, 200)

            if test['type'] == Tag.CODE:
                content = '\n'.join(test['content'])
                self.assertEqual(response.data.decode('utf-8'), f"<pre><code>{content}</code></pre>")

            if test['type'] == Tag.TEXT:
                self.assertIn(jsonify(test['content']).data.decode('UTF-8')[:-1], response.data.decode('UTF-8'))

            if test['type'] == Tag.IMAGE:
                content = ''.join(test['content'])
                self.assertEqual(response.data.decode('utf-8'), f"<img src='{content}'>")

            if test['type'] == Tag.COMMAND:
                result = subprocess.run(''.join(test['content']).split(), stdout=subprocess.PIPE, text=True)
                self.assertIn(result.stdout, response.data.decode('UTF-8'))

if __name__ == '__main__':
    unittest.main()
