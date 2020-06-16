package markdown

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/pgavlin/goldmark"
	"github.com/pgavlin/goldmark/ast"
	"github.com/pgavlin/goldmark/renderer"
	"github.com/pgavlin/goldmark/testutil"
	"github.com/pgavlin/goldmark/text"
	"github.com/pgavlin/goldmark/util"
)

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

			if !testutil.AssertSameStructure(t, sourceExpected, sourceActual, expected, actual, testutil.DefaultNodeAssertions()) {
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
