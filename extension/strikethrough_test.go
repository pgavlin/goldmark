package extension

import (
	"testing"

	"github.com/pgavlin/goldmark"
	"github.com/pgavlin/goldmark/renderer/html"
	"github.com/pgavlin/goldmark/testutil"
)

func TestStrikethrough(t *testing.T) {
	markdown := goldmark.New(
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
		goldmark.WithExtensions(
			Strikethrough,
		),
	)
	testutil.DoTestCaseFile(markdown, "_test/strikethrough.txt", t)
}
