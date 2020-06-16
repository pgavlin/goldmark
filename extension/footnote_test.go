package extension

import (
	"testing"

	"github.com/pgavlin/goldmark"
	"github.com/pgavlin/goldmark/renderer/html"
	"github.com/pgavlin/goldmark/testutil"
)

func TestFootnote(t *testing.T) {
	markdown := goldmark.New(
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
		goldmark.WithExtensions(
			Footnote,
		),
	)
	testutil.DoTestCaseFile(markdown, "_test/footnote.txt", t)
}
