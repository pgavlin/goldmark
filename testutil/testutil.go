package testutil

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/pgavlin/goldmark"
	"github.com/pgavlin/goldmark/ast"
	"github.com/pgavlin/goldmark/util"
)

// TestingT is a subset of the functionality provided by testing.T.
type TestingT interface {
	Logf(string, ...interface{})
	Skipf(string, ...interface{})
	Errorf(string, ...interface{})
	FailNow()
}

// MarkdownTestCase represents a test case.
type MarkdownTestCase struct {
	No          int
	Description string
	Markdown    string
	Expected    string
}

const attributeSeparator = "//- - - - - - - - -//"
const caseSeparator = "//= = = = = = = = = = = = = = = = = = = = = = = =//"

// DoTestCaseFile runs test cases in a given file.
func DoTestCaseFile(m goldmark.Markdown, filename string, t TestingT) {
	fp, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer fp.Close()

	scanner := bufio.NewScanner(fp)
	c := MarkdownTestCase{
		No:          -1,
		Description: "",
		Markdown:    "",
		Expected:    "",
	}
	cases := []MarkdownTestCase{}
	line := 0
	for scanner.Scan() {
		line++
		if util.IsBlank([]byte(scanner.Text())) {
			continue
		}
		header := scanner.Text()
		c.Description = ""
		if strings.Contains(header, ":") {
			parts := strings.Split(header, ":")
			c.No, err = strconv.Atoi(strings.TrimSpace(parts[0]))
			c.Description = strings.Join(parts[1:], ":")
		} else {
			c.No, err = strconv.Atoi(scanner.Text())
		}
		if err != nil {
			panic(fmt.Sprintf("%s: invalid case No at line %d", filename, line))
		}
		if !scanner.Scan() {
			panic(fmt.Sprintf("%s: invalid case at line %d", filename, line))
		}
		line++
		if scanner.Text() != attributeSeparator {
			panic(fmt.Sprintf("%s: invalid separator '%s' at line %d", filename, scanner.Text(), line))
		}
		buf := []string{}
		for scanner.Scan() {
			line++
			text := scanner.Text()
			if text == attributeSeparator {
				break
			}
			buf = append(buf, text)
		}
		c.Markdown = strings.Join(buf, "\n")
		buf = []string{}
		for scanner.Scan() {
			line++
			text := scanner.Text()
			if text == caseSeparator {
				break
			}
			buf = append(buf, text)
		}
		c.Expected = strings.Join(buf, "\n")
		cases = append(cases, c)
	}
	DoTestCases(m, cases, t)
}

// DoTestCases runs a set of test cases.
func DoTestCases(m goldmark.Markdown, cases []MarkdownTestCase, t TestingT) {
	for _, testCase := range cases {
		DoTestCase(m, testCase, t)
	}
}

// DoTestCase runs a test case.
func DoTestCase(m goldmark.Markdown, testCase MarkdownTestCase, t TestingT) {
	var ok bool
	var out bytes.Buffer
	defer func() {
		description := ""
		if len(testCase.Description) != 0 {
			description = ": " + testCase.Description
		}
		if err := recover(); err != nil {
			format := `============= case %d%s ================
Markdown:
-----------
%s

Expected:
----------
%s

Actual
---------
%v
%s
`
			t.Errorf(format, testCase.No, description, testCase.Markdown, testCase.Expected, err, debug.Stack())
		} else if !ok {
			format := `============= case %d%s ================
Markdown:
-----------
%s

Expected:
----------
%s

Actual
---------
%s
`
			t.Errorf(format, testCase.No, description, testCase.Markdown, testCase.Expected, out.Bytes())
		}
	}()

	if err := m.Convert([]byte(testCase.Markdown), &out); err != nil {
		panic(err)
	}
	ok = bytes.Equal(bytes.TrimSpace(out.Bytes()), bytes.TrimSpace([]byte(testCase.Expected)))
}

// AssertNodeFunc is used by AssertSameStructure to assert that two nodes are semantically identical.
type AssertNodeFunc func(t *testing.T, sourceA, sourceB []byte, a, b ast.Node) bool

func assertNodeNoop(t *testing.T, sa, sb []byte, a, b ast.Node) bool {
	return true
}

// NodeAssertions maps from node kinds to AssertNodeFunc.
type NodeAssertions map[ast.NodeKind]AssertNodeFunc

// Union returns a new set of node assertions that is the union of the two input sets.
func (na NodeAssertions) Union(other NodeAssertions) NodeAssertions {
	m := NodeAssertions{}
	for k, f := range na {
		m[k] = f
	}
	for k, f := range other {
		m[k] = f
	}
	return m
}

// DefaultNodeAssertions returns the default set of node assertions.
func DefaultNodeAssertions() NodeAssertions {
	return NodeAssertions{
		ast.KindAutoLink: func(t *testing.T, sa, sb []byte, a, b ast.Node) bool {
			na, nb := a.(*ast.AutoLink), b.(*ast.AutoLink)
			return assert.Equal(t, na.AutoLinkType, nb.AutoLinkType) &&
				assert.Equal(t, na.Protocol, nb.Protocol) &&
				assert.Equal(t, na.Label(sa), nb.Label(sb))
		},
		ast.KindEmphasis: func(t *testing.T, sa, sb []byte, a, b ast.Node) bool {
			na, nb := a.(*ast.Emphasis), b.(*ast.Emphasis)
			return assert.Equal(t, na.Level, nb.Level)
		},
		ast.KindFencedCodeBlock: func(t *testing.T, sa, sb []byte, a, b ast.Node) bool {
			na, nb := a.(*ast.FencedCodeBlock), b.(*ast.FencedCodeBlock)
			return assert.Equal(t, na.Language(sa), nb.Language(sb))
		},
		ast.KindHTMLBlock: func(t *testing.T, sa, sb []byte, a, b ast.Node) bool {
			return assert.Equal(t, a.Text(sa), b.Text(sb))
		},
		ast.KindHeading: func(t *testing.T, sa, sb []byte, a, b ast.Node) bool {
			na, nb := a.(*ast.Heading), b.(*ast.Heading)
			return assert.Equal(t, na.Level, nb.Level)
		},
		ast.KindImage: func(t *testing.T, sa, sb []byte, a, b ast.Node) bool {
			na, nb := a.(*ast.Image), b.(*ast.Image)
			return assert.Equal(t, na.Destination, nb.Destination) &&
				assert.Equal(t, na.Title, nb.Title)
		},
		ast.KindLink: func(t *testing.T, sa, sb []byte, a, b ast.Node) bool {
			na, nb := a.(*ast.Link), b.(*ast.Link)
			return assert.Equal(t, na.ReferenceType, nb.ReferenceType) &&
				AssertEqualBytes(t, na.Label, nb.Label) &&
				AssertEqualBytes(t, na.Destination, nb.Destination) &&
				AssertEqualBytes(t, na.Title, nb.Title)
		},
		ast.KindLinkReferenceDefinition: func(t *testing.T, sa, sb []byte, a, b ast.Node) bool {
			na, nb := a.(*ast.LinkReferenceDefinition), b.(*ast.LinkReferenceDefinition)
			return AssertEqualBytes(t, na.Label, nb.Label) &&
				AssertEqualBytes(t, na.Destination, nb.Destination) &&
				AssertEqualBytes(t, na.Title, nb.Title)
		},
		ast.KindList: func(t *testing.T, sa, sb []byte, a, b ast.Node) bool {
			na, nb := a.(*ast.List), b.(*ast.List)
			return assert.Equal(t, na.Marker, nb.Marker) &&
				assert.Equal(t, na.Start, nb.Start) &&
				assert.Equal(t, na.IsTight, nb.IsTight)
		},
		ast.KindRawHTML: func(t *testing.T, sa, sb []byte, a, b ast.Node) bool {
			na, nb := a.(*ast.RawHTML), b.(*ast.RawHTML)
			return assert.Equal(t, na.Text(sa), nb.Text(sb))
		},
		ast.KindString: func(t *testing.T, sa, sb []byte, a, b ast.Node) bool {
			na, nb := a.(*ast.String), b.(*ast.String)
			return assert.Equal(t, na.Value, nb.Value) &&
				assert.Equal(t, na.IsRaw(), nb.IsRaw())
		},
		ast.KindText: func(t *testing.T, sa, sb []byte, a, b ast.Node) bool {
			na, nb := a.(*ast.Text), b.(*ast.Text)
			return assert.Equal(t, na.Text(sa), nb.Text(sb)) &&
				assert.Equal(t, na.SoftLineBreak(), nb.SoftLineBreak()) &&
				assert.Equal(t, na.HardLineBreak(), nb.HardLineBreak()) &&
				assert.Equal(t, na.IsRaw(), nb.IsRaw())
		},
		ast.KindWhitespace: func(t *testing.T, sa, sb []byte, a, b ast.Node) bool {
			na, nb := a.(*ast.Whitespace), b.(*ast.Whitespace)
			return AssertEqualBytes(t, na.Segment.Value(sa), nb.Segment.Value(sb))
		},
		ast.KindBlockquote:    assertNodeNoop,
		ast.KindCodeBlock:     assertNodeNoop,
		ast.KindCodeSpan:      assertNodeNoop,
		ast.KindDocument:      assertNodeNoop,
		ast.KindListItem:      assertNodeNoop,
		ast.KindParagraph:     assertNodeNoop,
		ast.KindTextBlock:     assertNodeNoop,
		ast.KindThematicBreak: assertNodeNoop,
	}
}

// AssertEqualBytes asserts that the two input byte slices have the same length and contents.
func AssertEqualBytes(t *testing.T, a, b []byte) bool {
	if len(a) == 0 {
		a = nil
	}
	if len(b) == 0 {
		b = nil
	}
	return assert.Equal(t, a, b)
}

func assertSameStructure(t *testing.T, sa, sb []byte, a, b ast.Node, assertions NodeAssertions) bool {
	if !assert.Equal(t, a.Kind().String(), b.Kind().String()) {
		return false
	}

	assertFunc, ok := assertions[a.Kind()]
	if !ok {
		t.Logf("unexpected node kind %v", a.Kind())
	} else if !assertFunc(t, sa, sb, a, b) {
		return false
	}

	if !assert.Equal(t, a.ChildCount(), b.ChildCount()) {
		return false
	}

	for c, d := a.FirstChild(), b.FirstChild(); c != nil; c, d = c.NextSibling(), d.NextSibling() {
		if !assertSameStructure(t, sa, sb, c, d, assertions) {
			return false
		}
	}

	return true
}

// AssertSameStructure walks the ASTs rooted at a and b
func AssertSameStructure(t *testing.T, sourceA, sourceB []byte, a, b ast.Node, assertions NodeAssertions) bool {
	return assertSameStructure(t, sourceA, sourceB, a, b, assertions)
}
