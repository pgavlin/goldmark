package ast

import (
	"fmt"
	gast "github.com/yuin/goldmark/ast"
	"io"
)

// A TaskCheckBox struct represents a checkbox of a task list.
type TaskCheckBox struct {
	gast.BaseInline
	IsChecked bool
}

// Dump implements Node.Dump.
func (n *TaskCheckBox) Dump(w io.Writer, source []byte, level int) {
	m := map[string]string{
		"Checked": fmt.Sprintf("%v", n.IsChecked),
	}
	gast.DumpHelper(w, n, source, level, m, nil)
}

// KindTaskCheckBox is a NodeKind of the TaskCheckBox node.
var KindTaskCheckBox = gast.NewNodeKind("TaskCheckBox")

// Kind implements Node.Kind.
func (n *TaskCheckBox) Kind() gast.NodeKind {
	return KindTaskCheckBox
}

// NewTaskCheckBox returns a new TaskCheckBox node.
func NewTaskCheckBox(checked bool) *TaskCheckBox {
	return &TaskCheckBox{
		IsChecked: checked,
	}
}
