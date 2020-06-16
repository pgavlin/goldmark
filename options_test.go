package goldmark_test

import (
	"testing"

	. "github.com/pgavlin/goldmark"
	"github.com/pgavlin/goldmark/parser"
	"github.com/pgavlin/goldmark/testutil"
)

func TestAttributeAndAutoHeadingID(t *testing.T) {
	markdown := New(
		WithParserOptions(
			parser.WithAttribute(),
			parser.WithAutoHeadingID(),
		),
	)
	testutil.DoTestCaseFile(markdown, "_test/options.txt", t)
}
