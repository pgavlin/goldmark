package extension

import (
	"testing"

	"github.com/pgavlin/goldmark"
	"github.com/pgavlin/goldmark/renderer/html"
	"github.com/pgavlin/goldmark/testutil"
)

func TestDefinitionList(t *testing.T) {
	markdown := goldmark.New(
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
		goldmark.WithExtensions(
			DefinitionList,
		),
	)
	testutil.DoTestCaseFile(markdown, "_test/definition_list.txt", t)
}
