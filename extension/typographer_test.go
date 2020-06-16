package extension

import (
	"testing"

	"github.com/pgavlin/goldmark"
	"github.com/pgavlin/goldmark/renderer/html"
	"github.com/pgavlin/goldmark/testutil"
)

func TestTypographer(t *testing.T) {
	markdown := goldmark.New(
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
		goldmark.WithExtensions(
			Typographer,
		),
	)
	testutil.DoTestCaseFile(markdown, "_test/typographer.txt", t)
}
