// Package ast defines AST nodes that represents extension's elements
package ast

import (
	gast "github.com/pgavlin/goldmark/ast"
	"io"
)

// A Strikethrough struct represents a strikethrough of GFM text.
type Strikethrough struct {
	gast.BaseInline
}

// Dump implements Node.Dump.
func (n *Strikethrough) Dump(w io.Writer, source []byte, level int) {
	gast.DumpHelper(w, n, source, level, nil, nil)
}

// KindStrikethrough is a NodeKind of the Strikethrough node.
var KindStrikethrough = gast.NewNodeKind("Strikethrough")

// Kind implements Node.Kind.
func (n *Strikethrough) Kind() gast.NodeKind {
	return KindStrikethrough
}

// NewStrikethrough returns a new Strikethrough node.
func NewStrikethrough() *Strikethrough {
	return &Strikethrough{}
}
