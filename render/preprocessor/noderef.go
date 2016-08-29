package preprocessor

import "text/template/parse"

// ThunkFunc represents a thunk over a node operation
type ThunkFunc func() func(i interface{}) interface{}

// NodeRef handles state for calls to `noderef` from desugared intra-module
// value references.  A NodeRef is essentially a thunk to defer intra-module
// referencing.
type NodeRef struct {
	id    string
	funcs map[string]ThunkFunc
}

// RefNode takes a parse node and returns a noderef with thunks for each xref
func RefNode(node *parse.Node) (*NodeRef, error) {

}
