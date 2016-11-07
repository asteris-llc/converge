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
	fmt.Printf("Getting peer branches for: %s\n", meta.ID)
	if !IsConditional(meta) {
		fmt.Printf("\t not a conditional node\n")
		return
	}
	kind, ok := meta.LookupMetadata(MetaType)
	if !ok {
		fmt.Printf("\t No type information for node\n")
		return
	}
	fmt.Printf("\t type: %s\n", kind)
	switch kind {
	case NodeCatSwitch:
		fmt.Printf("\t skipping switch root node\n")
		return
	case NodeCatResource:
		parent, ok := g.GetParent(meta.ID)
		fmt.Printf("\t found a resource node, deferring to parent: %s\n", parent)
		if !ok {
			fmt.Printf("\t\t no parent node found\n")
			return
		}
		return PeerBranches(g, parent)
	}
	fmt.Printf("\t found a branch node; getting embedded peers list...\n")
	peerStrsI, ok := meta.LookupMetadata(MetaPeers)
	if !ok {
		fmt.Printf("\t no peers in metadata\n")
		return
	}
	fmt.Printf("\t trying to convert peer list to a string slice...\n")
	peerStrs, ok := peerStrsI.([]string)
	if !ok {
		fmt.Printf("\t peers are not a string slice, returning\n")
		return
	}
	fmt.Printf("\t raw peer list: %v\n", peerStrs)
	fmt.Printf("\t trying to get parent id...\n")
	parentID, ok := g.GetParentID(meta.ID)
	if !ok {
		fmt.Printf("\t no parent for current node, returning")
		return
	}
	fmt.Printf("\t my parent: %s\n", parentID)
	fmt.Printf("\t getting peer nodes...\n")
	for _, peer := range peerStrs {
		peerPath := graph.ID(parentID, peer)
		fmt.Printf("\t\t %s\n", peerPath)
		if meta, ok := g.Get(peerPath); ok {
			fmt.Printf("\t\t found in graph... adding\n")
			out = append(out, meta)
		} else {
			fmt.Printf("\t\t not found in graph, skipping\n")
		}
	}
	fmt.Printf("finished processing peer branches: got %d peers\n", len(out))
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
