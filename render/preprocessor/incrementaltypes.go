package preprocessor

import "github.com/asteris-llc/converge/graph"

// NodeThunk represents a lazy list of thunks with the final element being the
// resource.TaskStatus returned from calling Check
type NodeThunk func() (interface{}, interface{})

// LoadThunk creates a thunk for a given node
func LoadThunk(g *graph.Graph, id string) (Incremental, error) {

}
