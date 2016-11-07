package conditional

import (
	"errors"
	"fmt"
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

	// MetaType contains the type of the underlying node
	MetaType = "conditional-resource-type"
)

// A NodeCategory represents the type of a node
type NodeCategory string

const (
	// NodeCatResource represents a NodeCat that is a node inside of a branch
	NodeCatResource = "resource"
	// NodeCatBranch represents a NodeCat that is a branch (case statement)
	NodeCatBranch = "branch"
	// NodeCatSwitch represents a conditional container type (switch statement)
	NodeCatSwitch = "switch"
)

var (
	// ErrUnrendered is returned when attempting to evaluate truthiness of an
	// unrendered predicate.
	ErrUnrendered = errors.New("cannot evaluate an unrendered predicate")
)

// PeerNodes returns a list of graph nodes that are part of the same switch
// statement and branch as the current node.
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

// PeerBranches returns the branches nodes that are peers to the current node.
func PeerBranches(g *graph.Graph, meta *node.Node) (out []*node.Node) {
	if !IsConditional(meta) {
		return
	}
	kind, ok := meta.LookupMetadata(MetaType)
	if !ok {
		return
	}
	switch kind {
	case NodeCatSwitch:
		return
	case NodeCatResource:
		parent, ok := g.GetParent(meta.ID)
		if !ok {
			return
		}
		return PeerBranches(g, parent)
	}
	peerStrsI, ok := meta.LookupMetadata(MetaPeers)
	if !ok {
		return
	}
	peerStrs, ok := peerStrsI.([]string)
	if !ok {
		return
	}
	parentID, ok := g.GetParentID(meta.ID)
	if !ok {
		return
	}
	for _, peer := range peerStrs {
		peerPath := graph.ID(parentID, peer)
		if meta, ok := g.Get(peerPath); ok {
			out = append(out, meta)
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
		return "", errors.New("\t predicate required for conditional node")
	}
	toRender := unrendered.(string)
	result, err := renderFunc(meta.ID, toRender)
	if err != nil {
		return "", err
	}
	result, err = renderFunc(meta.ID, fmt.Sprintf("{{ %s }}", result))
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

// ShouldEvaluate returns true if the node is the first of it's peers that is
// true.
func ShouldEvaluate(g *graph.Graph, meta *node.Node) (bool, error) {
	myBranch, ok := meta.LookupMetadata(MetaBranchName)
	if !ok {
		return false, errors.New("invalid branch identifier for " + meta.ID)
	}

	fmt.Printf("calling should evaluate branch %s in %v\n", myBranch, PeerBranches(g, meta))

	for _, node := range PeerBranches(g, meta) {
		nodeBranch, ok := node.LookupMetadata(MetaBranchName)
		if !ok {
			return false, errors.New("invalid branch identifier for " + node.ID)
		}
		if myBranch == nodeBranch {
			break
		}
		if ok, err := IsTrue(node); ok || err != nil {
			return false, err
		}
	}
	res, err := IsTrue(meta)
	return res, err
}
