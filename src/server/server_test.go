package server

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServer(t *testing.T) {
	t.Run("With preCommands", func(t *testing.T) {
		commands := filepath.Join("..", "testdata", "commands.txt")

		s, err := NewServer(slog.New(slog.NewTextHandler(os.Stdout, nil)), commands)
		require.NoError(t, err)

		slideContent, err := s.SplitContent(commands)
		require.NoError(t, err)
		assert.Equal(t, []string{
			"TEST=$(echo 345)",
			"TEST2=123",
			"---",
			"this is a [presentation](http://google.com)",
			"$ echo aaa && sleep 2 && echo bbb",
			"$ echo $TEST",
			"```python",
			"def main():",
			"    print(\"hello world\")",
			"```",
			"![hyperion](https://upload.wikimedia.org/wikipedia/en/7/73/Hyperion_cover.jpg)",
			"$ echo hello world",
			"$ watch -n 1 date",
			"this is `some` text",
			"$ ls /",
			"$ echo {} | jq",
			"$ adsadads",
			"$ ls -R /",
		}, slideContent)

		assert.Equal(t, []string{"TEST=$(echo 345)", "TEST2=123"}, s.GetPreCommands())

		assert.Equal(t, slide{0, "this is a <a href=\"http://google.com\" target=\"_blank\" rel=\"noopener noreferrer\">presentation</a>", slideTypePlain}, s.GetSlide(0))
		assert.Equal(t, slide{1, "echo aaa && sleep 2 && echo bbb", slideTypeCommand}, s.GetSlide(1))
		assert.Equal(t, slide{3, "<pre class='language-python'><code>\ndef main():\n    print(\"hello world\")\n</code></pre>", slideTypeCodeblock}, s.GetSlide(3))
		assert.Equal(t, slide{4, "<img src=\"https://upload.wikimedia.org/wikipedia/en/7/73/Hyperion_cover.jpg\" alt=\"hyperion\">", slideTypePlain}, s.GetSlide(4))

		assert.Equal(t, 12, s.GetSlideCount())
	})

	t.Run("With preCommands", func(t *testing.T) {
		commands := filepath.Join("..", "testdata", "no-pre-commands.txt")

		s, err := NewServer(slog.New(slog.NewTextHandler(os.Stdout, nil)), commands)
		require.NoError(t, err)

		slideContent, err := s.SplitContent(commands)
		require.NoError(t, err)
		assert.Equal(t, []string{
			"this is a [presentation](http://google.com)",
			"$ echo aaa && sleep 2 && echo bbb",
			"$ echo $TEST",
			"```python",
			"def main():",
			"    print(\"hello world\")",
			"```",
			"![hyperion](https://upload.wikimedia.org/wikipedia/en/7/73/Hyperion_cover.jpg)",
			"$ echo hello world",
			"$ watch -n 1 date",
			"this is `some` text",
			"$ ls /",
			"$ echo {} | jq",
			"$ adsadads",
			"$ ls -R /",
		}, slideContent)

		assert.Equal(t, []string{}, s.GetPreCommands())

		assert.Equal(t, slide{0, "this is a <a href=\"http://google.com\" target=\"_blank\" rel=\"noopener noreferrer\">presentation</a>", slideTypePlain}, s.GetSlide(0))
		assert.Equal(t, slide{1, "echo aaa && sleep 2 && echo bbb", slideTypeCommand}, s.GetSlide(1))
		assert.Equal(t, slide{3, "<pre class='language-python'><code>\ndef main():\n    print(\"hello world\")\n</code></pre>", slideTypeCodeblock}, s.GetSlide(3))
		assert.Equal(t, slide{4, "<img src=\"https://upload.wikimedia.org/wikipedia/en/7/73/Hyperion_cover.jpg\" alt=\"hyperion\">", slideTypePlain}, s.GetSlide(4))

		assert.Equal(t, 12, s.GetSlideCount())
	})
}

func TestParsePlainSlide(t *testing.T) {
	testCases := []struct {
		input, output string
	}{
		{
			"i am a sentence",
			"i am a sentence",
		},
		{
			"i am a `sentence`",
			"i am a <code>sentence</code>",
		},
		{
			"i am a `sentence` too",
			"i am a <code>sentence</code> too",
		},
		{
			"i am a [apple](sentence)",
			"i am a <a href=\"sentence\" target=\"_blank\" rel=\"noopener noreferrer\">apple</a>",
		},
		{
			"i am a [](sentence)",
			"i am a <a href=\"sentence\" target=\"_blank\" rel=\"noopener noreferrer\"></a>",
		},
		{
			"i am a []()",
			"i am a []()",
		},
		{
			// image only slides are allowed but not ones with text
			"i am a ![apple](sentence)",
			"i am a ![apple](sentence)",
		},
		{
			// image only slides are allowed but not ones with text
			"i am a ![](sentence)",
			"i am a ![](sentence)",
		},
		{
			// image only slides are allowed but not ones with text
			"i am a ![]()",
			"i am a ![]()",
		},
		{
			"![apple](sentence)",
			"<img src=\"sentence\" alt=\"apple\">",
		},
		{
			"![](sentence)",
			"<img src=\"sentence\" alt=\"\">",
		},
		{
			"![]()",
			"![]()",
		},
	}

	for i := range testCases {
		test := testCases[i]
		expected := test.output
		actual := parsePlainSlide(test.input)
		assert.Equal(t, expected, actual)
	}
}
