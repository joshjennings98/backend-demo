package server

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/joshjennings98/backend-demo/server/types"
)

func TestNewServer(t *testing.T) {
	t.Run("With preCommands", func(t *testing.T) {
		commands := filepath.Join("..", "testdata", "commands.txt")

		s, err := NewServer(slog.New(slog.NewTextHandler(os.Stdout, nil)), commands)
		require.NoError(t, err)

		preCommands, slideContent, err := s.SplitContent(commands)
		require.NoError(t, err)
		assert.Equal(t, []string{
			"TEST=$(echo 345) # comment",
			"TEST2=123",
			"# comment won't get run",
			"echo test",
		}, preCommands)
		assert.Equal(t, []string{
			"this is a [presentation](http://google.com)",
			"$ echo aaa && sleep 2 && echo bbb",
			"$ echo $TEST",
			"```python\ndef main():\n    print(\"hello world\")\n```",
			"![hyperion](https://upload.wikimedia.org/wikipedia/en/7/73/Hyperion_cover.jpg)",
			"<iframe src=\"https://www.google.com\"></iframe>",
			"$ echo hello world",
			"$ watch -n 1 date",
			"this is `some` text",
			"$ ls /",
			"$ echo {} | jq",
			"$ adsadads",
			"$ ls -R /",
		}, slideContent)

		assert.Equal(t, []string{"TEST=$(echo 345) # comment", "TEST2=123", "# comment won't get run", "echo test"}, s.GetPreCommands())

		assert.Equal(t, types.Slide{ID: 0, Content: "<p>this is a <a href=\"http://google.com\">presentation</a></p>\n", SlideType: types.SlideTypePlain}, s.GetSlide(0))
		assert.Equal(t, types.Slide{ID: 1, Content: "echo aaa && sleep 2 && echo bbb", SlideType: types.SlideTypeCommand}, s.GetSlide(1))
		assert.Equal(t, types.Slide{ID: 3, Content: "<pre><code class=\"language-python\">def main():\n    print(&quot;hello world&quot;)\n</code></pre>\n", SlideType: types.SlideTypeCodeblock}, s.GetSlide(3))
		assert.Equal(t, types.Slide{ID: 4, Content: "<p><img src=\"https://upload.wikimedia.org/wikipedia/en/7/73/Hyperion_cover.jpg\" alt=\"hyperion\" /></p>\n", SlideType: types.SlideTypePlain}, s.GetSlide(4))

		assert.Equal(t, 13, s.GetSlideCount())
	})

	t.Run("Without preCommands", func(t *testing.T) {
		commands := filepath.Join("..", "testdata", "no-pre-commands.txt")

		s, err := NewServer(slog.New(slog.NewTextHandler(os.Stdout, nil)), commands)
		require.NoError(t, err)

		preCommands, slideContent, err := s.SplitContent(commands)
		require.NoError(t, err)
		assert.Equal(t, []string{
			"this is a [presentation](http://google.com)",
			"$ echo aaa && sleep 2 && echo bbb",
			"$ echo $TEST",
			"```python\ndef main():\n    print(\"hello world\")\n```",
			"![hyperion](https://upload.wikimedia.org/wikipedia/en/7/73/Hyperion_cover.jpg)",
			"<iframe>https://www.google.com</iframe>",
			"$ echo hello world",
			"$ watch -n 1 date",
			"this is `some` text",
			"$ ls /",
			"$ echo {} | jq",
			"$ adsadads",
			"$ ls -R /",
		}, slideContent)

		assert.Empty(t, preCommands)
		assert.Empty(t, s.GetPreCommands())

		assert.Equal(t, types.Slide{ID: 0, Content: "<p>this is a <a href=\"http://google.com\">presentation</a></p>\n", SlideType: types.SlideTypePlain}, s.GetSlide(0))
		assert.Equal(t, types.Slide{ID: 1, Content: "echo aaa && sleep 2 && echo bbb", SlideType: types.SlideTypeCommand}, s.GetSlide(1))
		assert.Equal(t, types.Slide{ID: 3, Content: "<pre><code class=\"language-python\">def main():\n    print(&quot;hello world&quot;)\n</code></pre>\n", SlideType: types.SlideTypeCodeblock}, s.GetSlide(3))
		assert.Equal(t, types.Slide{ID: 4, Content: "<p><img src=\"https://upload.wikimedia.org/wikipedia/en/7/73/Hyperion_cover.jpg\" alt=\"hyperion\" /></p>\n", SlideType: types.SlideTypePlain}, s.GetSlide(4))

		assert.Equal(t, 13, s.GetSlideCount())
	})

	t.Run("Multiple --- in command file", func(t *testing.T) {
		commands := filepath.Join("..", "testdata", "broken-pre-commands.txt")

		_, err := NewServer(slog.New(slog.NewTextHandler(os.Stdout, nil)), commands)
		assert.ErrorContains(t, err, "more than one '---' was found")
	})

}

func TestInitialise(t *testing.T) {
	t.Run("Normal", func(t *testing.T) {
		commands := filepath.Join("..", "testdata", "commands.txt")

		s, err := NewServer(slog.New(slog.NewTextHandler(os.Stdout, nil)), commands)
		require.NoError(t, err)

		err = s.Initialise(context.Background())
		assert.NoError(t, err)
	})

	t.Run("Bad pre command", func(t *testing.T) {
		commands := filepath.Join("..", "testdata", "bad-pre-commands.txt")

		s, err := NewServer(slog.New(slog.NewTextHandler(os.Stdout, nil)), commands)
		require.NoError(t, err)

		err = s.Initialise(context.Background())
		assert.ErrorContains(t, err, "error executing command")
	})

	t.Run("Bad pre command", func(t *testing.T) {
		commands := filepath.Join("..", "testdata", "bad-pre-commands1.txt")

		s, err := NewServer(slog.New(slog.NewTextHandler(os.Stdout, nil)), commands)
		require.NoError(t, err)

		err = s.Initialise(context.Background())
		assert.ErrorContains(t, err, "error executing command")
	})

	t.Run("Cancel context", func(t *testing.T) {
		commands := filepath.Join("..", "testdata", "commands.txt")

		s, err := NewServer(slog.New(slog.NewTextHandler(os.Stdout, nil)), commands)
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err = s.Initialise(ctx)
		assert.ErrorContains(t, err, "canceled")
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
			"i am a <a href=\"sentence\">apple</a>",
		},
		{
			"i am a [](sentence)",
			"i am a <a href=\"sentence\"></a>",
		},
		{
			"i am a []()",
			"i am a <a href=\"\"></a>",
		},
		{
			"i am a ![apple](sentence)",
			"i am a <img src=\"sentence\" alt=\"apple\" />",
		},
		{
			"i am a ![](sentence)",
			"i am a <img src=\"sentence\" alt=\"\" />",
		},
		{
			"i am a ![]()",
			"i am a <img src=\"\" alt=\"\" />",
		},
		{
			"![apple](sentence)",
			"<img src=\"sentence\" alt=\"apple\" />",
		},
		{
			"![](sentence)",
			"<img src=\"sentence\" alt=\"\" />",
		},
		{
			"![]()",
			"<img src=\"\" alt=\"\" />",
		},
	}

	for i := range testCases {
		test := testCases[i]
		expected := fmt.Sprintf("<p>%v</p>\n", test.output)
		actual := parseSlide(test.input)
		assert.Equal(t, expected, actual)
	}
}
