package extension

import (
	"testing"

	"github.com/pgavlin/goldmark"
	"github.com/pgavlin/goldmark/renderer/html"
	"github.com/pgavlin/goldmark/testutil"
)

func TestTaskList(t *testing.T) {
	markdown := goldmark.New(
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
		goldmark.WithExtensions(
			TaskList,
		),
	)
	testutil.DoTestCaseFile(markdown, "_test/tasklist.txt", t)
}
