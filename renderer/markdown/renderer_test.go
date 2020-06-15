package markdown

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

func assertEqualBytes(t *testing.T, a, b []byte) bool {
	if len(a) == 0 {
		a = nil
	}
	if len(b) == 0 {
		b = nil
	}
	return assert.Equal(t, a, b)
}

func assertSameStructure(t *testing.T, sa, sb []byte, a, b ast.Node) bool {
	if !assert.Equal(t, a.Kind().String(), b.Kind().String()) {
		return false
	}

	var ok bool
	switch a.Kind() {
	case ast.KindAutoLink:
		na, nb := a.(*ast.AutoLink), b.(*ast.AutoLink)
		ok = assert.Equal(t, na.AutoLinkType, nb.AutoLinkType) &&
			assert.Equal(t, na.Protocol, nb.Protocol) &&
			assert.Equal(t, na.Label(sa), nb.Label(sb))
	case ast.KindEmphasis:
		na, nb := a.(*ast.Emphasis), b.(*ast.Emphasis)
		ok = assert.Equal(t, na.Level, nb.Level)
	case ast.KindFencedCodeBlock:
		na, nb := a.(*ast.FencedCodeBlock), b.(*ast.FencedCodeBlock)
		ok = assert.Equal(t, na.Language(sa), nb.Language(sb))
	case ast.KindHTMLBlock:
		assert.Equal(t, a.Text(sa), b.Text(sb))
	case ast.KindHeading:
		na, nb := a.(*ast.Heading), b.(*ast.Heading)
		ok = assert.Equal(t, na.Level, nb.Level)
	case ast.KindImage:
		na, nb := a.(*ast.Image), b.(*ast.Image)
		ok = assert.Equal(t, na.Destination, nb.Destination) &&
			assert.Equal(t, na.Title, nb.Title)
	case ast.KindLink:
		na, nb := a.(*ast.Link), b.(*ast.Link)
		ok = assert.Equal(t, na.ReferenceType, nb.ReferenceType) &&
			assertEqualBytes(t, na.Label, nb.Label) &&
			assertEqualBytes(t, na.Destination, nb.Destination) &&
			assertEqualBytes(t, na.Title, nb.Title)
	case ast.KindLinkReferenceDefinition:
		na, nb := a.(*ast.LinkReferenceDefinition), b.(*ast.LinkReferenceDefinition)
		ok = assertEqualBytes(t, na.Label, nb.Label) &&
			assertEqualBytes(t, na.Destination, nb.Destination) &&
			assertEqualBytes(t, na.Title, nb.Title)
	case ast.KindList:
		na, nb := a.(*ast.List), b.(*ast.List)
		ok = assert.Equal(t, na.Marker, nb.Marker) &&
			assert.Equal(t, na.Start, nb.Start) &&
			assert.Equal(t, na.IsTight, nb.IsTight)
	case ast.KindRawHTML:
		na, nb := a.(*ast.RawHTML), b.(*ast.RawHTML)
		ok = assert.Equal(t, na.Text(sa), nb.Text(sb))
	case ast.KindString:
		na, nb := a.(*ast.String), b.(*ast.String)
		ok = assert.Equal(t, na.Value, nb.Value) &&
			assert.Equal(t, na.IsRaw(), nb.IsRaw())
	case ast.KindText:
		na, nb := a.(*ast.Text), b.(*ast.Text)
		ok = assert.Equal(t, na.Text(sa), nb.Text(sb)) &&
			assert.Equal(t, na.SoftLineBreak(), nb.SoftLineBreak()) &&
			assert.Equal(t, na.HardLineBreak(), nb.HardLineBreak()) &&
			assert.Equal(t, na.IsRaw(), nb.IsRaw())
	case ast.KindWhitespace:
		na, nb := a.(*ast.Whitespace), b.(*ast.Whitespace)
		ok = assertEqualBytes(t, na.Segment.Value(sa), nb.Segment.Value(sb))
	case ast.KindBlockquote, ast.KindCodeBlock, ast.KindCodeSpan, ast.KindDocument, ast.KindListItem, ast.KindParagraph,
		ast.KindTextBlock, ast.KindThematicBreak:
		ok = true

		// Nothing extra to check
	default:
		t.Logf("unexpected node kind %v", a.Kind())
	}
	if !ok {
		return false
	}

	if !assert.Equal(t, a.ChildCount(), b.ChildCount()) {
		return false
	}

	for c, d := a.FirstChild(), b.FirstChild(); c != nil; c, d = c.NextSibling(), d.NextSibling() {
		if !assertSameStructure(t, sa, sb, c, d) {
			return false
		}
	}

	return true
}

type commonmarkSpecTestCase struct {
	Markdown string `json:"markdown"`
	Example  int    `json:"example"`
}

func readTestCases(path string) ([]commonmarkSpecTestCase, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var testCases []commonmarkSpecTestCase
	if err := json.NewDecoder(f).Decode(&testCases); err != nil {
		return nil, err
	}
	return testCases, nil
}

func sdump(node ast.Node, source []byte) string {
	var buf bytes.Buffer
	node.Dump(&buf, source, 0)
	return buf.String()
}

func TestSpec(t *testing.T) {
	testCases, err := readTestCases("../../_test/spec.json")
	if err != nil {
		t.Fatalf("failed to read test cases from spec.json: %v", err)
	}

	for _, c := range testCases {
		if caseToRun != -1 && c.Example != caseToRun {
			continue
		}

		t.Run(fmt.Sprintf("case %d", c.Example), func(t *testing.T) {
			sourceExpected := []byte(c.Markdown)
			parser := goldmark.DefaultParser()
			expected := parser.Parse(text.NewReader(sourceExpected))

			var buf bytes.Buffer
			renderer := renderer.NewRenderer(renderer.WithNodeRenderers(util.Prioritized(&Renderer{}, 100)))
			err := renderer.Render(&buf, sourceExpected, expected)
			if !assert.NoError(t, err) {
				t.Fatal()
			}
			sourceActual := buf.Bytes()
			actual := parser.Parse(text.NewReader(sourceActual))

			if !assertSameStructure(t, sourceExpected, sourceActual, expected, actual) {
				t.Logf("case %d:", c.Example)

				t.Logf("expected: %q", string(sourceExpected))
				t.Logf("%s", sdump(expected, sourceExpected))

				t.Logf("actual: %q", string(sourceActual))
				t.Logf("%s", sdump(actual, sourceActual))
			}
		})
	}
}

var caseToRun int

func TestMain(m *testing.M) {
	flag.IntVar(&caseToRun, "case", -1, "a single case to run in TestSpec")
	flag.Parse()

	os.Exit(m.Run())
}
