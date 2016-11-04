package conditional

import (
	"errors"
	"strings"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/graph/node"
)

const (
	// MetaSwitchName is the metadata key for accessing the name of the switch
	// that contains this branch.
	MetaSwitchName = "conditional-switch-name"
	// MetaBranchName defines the name for the branch
	MetaBranchName = "conditional-name"
	// MetaUnrenderedPredicate contains the predicate string before being rendered
	MetaUnrenderedPredicate = "conditional-predicate-raw"
	// MetaRenderedPredicate contains the predicate string after rendering
	MetaRenderedPredicate = "conditional-predicate-rendered"
	// MetaConditionalName contains the name of the branch containing the node
	MetaConditionalName = "conditional-name"
	// MetaPeers contains the the current branch and all it's peers in order
	MetaPeers = "conditional-peers"
	// MetaPredicate contains the cached result of rendered predicate evaluation
	MetaPredicate = "conditional-predicate-results"
)

var (
	// ErrUnrendered is returned when attempting to evaluate truthiness of an
	// unrendered predicate.
	ErrUnrendered = errors.New("cannot evaluate an unrendered predicate")
)

// PeerNodes returns a list of graph nodes that are peers to the current
// conditional node
func PeerNodes(g *graph.Graph, meta *node.Node) (out []*node.Node) {
	peers, ok := meta.LookupMetadata(MetaPeers)
	if !ok {
		return
	}
	parent, ok := g.GetParentID(meta.ID)
	if !ok {
		return
	}
	for _, peer := range peers.([]string) {
		if peerNode, ok := g.Get(graph.ID(parent, peer)); ok {
			out = append(out, peerNode)
		}
	}
	return
}

// IsConditional returns true if the graph is conditional
func IsConditional(meta *node.Node) bool {
	_, ok := meta.LookupMetadata("conditional-name")
	return ok
}

// RenderPredicate will attempt to render the predicate if it's not rendered,
// and return it.  It takes a renderFunc in order to prevent circular imports
// when being used by render.
func RenderPredicate(meta *node.Node, renderFunc func(string, string) (string, error)) (string, error) {
	rendered, ok := meta.LookupMetadata(MetaRenderedPredicate)
	if ok {
		return rendered.(string), nil
	}
	unrendered, ok := meta.LookupMetadata(MetaUnrenderedPredicate)
	if !ok {
		return "", errors.New("predicate required for conditional node")
	}
	result, err := renderFunc(meta.ID, unrendered.(string))
	if err != nil {
		return "", err
	}
	meta.AddMetadata(MetaRenderedPredicate, result)
	return result, nil
}

// IsTrue returns true if the predicate is true, and false otherwise.
func IsTrue(meta *node.Node) (bool, error) {
	var truth bool
	if val, ok := meta.LookupMetadata(MetaPredicate); ok {
		return val.(bool), nil
	}
	rendered, ok := meta.LookupMetadata(MetaRenderedPredicate)
	if !ok {
		return false, ErrUnrendered
	}
	switch strings.ToLower(rendered.(string)) {
	case "t", "true":
		truth = true
	default:
		truth = false
	}
	meta.AddMetadata(MetaPredicate, truth)
	return truth, nil
}
