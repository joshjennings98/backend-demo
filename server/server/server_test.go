package server

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/joshjennings98/backend-demo/server/v2/types"
)

func TestNewServer(t *testing.T) {
	t.Run("Basic presentation", func(t *testing.T) {
		commands := filepath.Join("..", "testdata", "commands.txt")

		s, err := NewServer(slog.New(slog.NewTextHandler(os.Stdout, nil)), 0, commands)
		require.NoError(t, err)

		slideContent, err := s.SplitContent(commands)
		require.NoError(t, err)
		assert.Equal(t, []string{
			"this is a [presentation](http://google.com)",
			"$ echo aaa && sleep 2 && echo bbb",
			"$ TEST=345\n$ echo $TEST",
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
			"$! echo \"visible setup line\"\n$ echo \"main command\"",
		}, slideContent)

		slide0, err := s.GetSlide(0)
		require.NoError(t, err)
		assert.Equal(t, types.Slide{
			ID:             0,
			Content:        "<p>this is a <a href=\"http://google.com\">presentation</a></p>\n",
			ExecuteContent: nil,
			SlideType:      types.SlideTypePlain,
		}, slide0)

		slide1, err := s.GetSlide(1)
		require.NoError(t, err)
		assert.Equal(t, types.Slide{
			ID:             1,
			Content:        "echo aaa && sleep 2 && echo bbb",
			ExecuteContent: []string{"echo aaa && sleep 2 && echo bbb"},
			SlideType:      types.SlideTypeCommand,
		}, slide1)

		slide2, err := s.GetSlide(2)
		require.NoError(t, err)
		assert.Equal(t, types.Slide{
			ID:             2,
			Content:        "echo $TEST",
			ExecuteContent: []string{"TEST=345", "echo $TEST"},
			SlideType:      types.SlideTypeCommand,
		}, slide2)

		slide3, err := s.GetSlide(3)
		require.NoError(t, err)
		assert.Equal(t, types.Slide{
			ID:             3,
			Content:        "<pre><code class=\"language-python\">def main():\n    print(&quot;hello world&quot;)\n</code></pre>\n",
			ExecuteContent: nil,
			SlideType:      types.SlideTypeCodeblock,
		}, slide3)

		slide4, err := s.GetSlide(4)
		require.NoError(t, err)
		assert.Equal(t, types.Slide{
			ID:             4,
			Content:        "<p><img src=\"https://upload.wikimedia.org/wikipedia/en/7/73/Hyperion_cover.jpg\" alt=\"hyperion\" /></p>\n",
			ExecuteContent: nil,
			SlideType:      types.SlideTypePlain,
		}, slide4)

		slide13, err := s.GetSlide(13)
		require.NoError(t, err)
		assert.Equal(t, types.Slide{
			ID:             13,
			Content:        "echo \"visible setup line\"\necho \"main command\"",
			ExecuteContent: []string{"echo \"visible setup line\"", "echo \"main command\""},
			SlideType:      types.SlideTypeCommand,
		}, slide13)

		_, err = s.GetSlide(100)
		assert.ErrorIs(t, err, ErrSlideIndexOutOfBounds)
		_, err = s.GetSlide(-1)
		assert.ErrorIs(t, err, ErrSlideIndexOutOfBounds)

		assert.Equal(t, 14, s.GetSlideCount())
	})
}

func TestParseCommandSlide(t *testing.T) {
	testCases := []struct {
		name           string
		input          string
		displayContent []string
		executeContent []string
	}{
		{
			name:           "single command",
			input:          "$ echo hello",
			displayContent: []string{"echo hello"},
			executeContent: []string{"echo hello"},
		},
		{
			name:           "hidden setup + visible main",
			input:          "$ export FOO=bar\n$ echo $FOO",
			displayContent: []string{"echo $FOO"},
			executeContent: []string{"export FOO=bar", "echo $FOO"},
		},
		{
			name:           "visible setup with $!",
			input:          "$! echo setup\n$ echo main",
			displayContent: []string{"echo setup", "echo main"},
			executeContent: []string{"echo setup", "echo main"},
		},
		{
			name:           "multiple hidden + one visible",
			input:          "$ cmd1\n$ cmd2\n$ cmd3",
			displayContent: []string{"cmd3"},
			executeContent: []string{"cmd1", "cmd2", "cmd3"},
		},
		{
			name:           "multiple $! lines",
			input:          "$! line1\n$! line2\n$ line3",
			displayContent: []string{"line1", "line2", "line3"},
			executeContent: []string{"line1", "line2", "line3"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			display, execute := parseCommandSlide(tc.input)
			assert.Equal(t, tc.displayContent, display)
			assert.Equal(t, tc.executeContent, execute)
		})
	}
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
