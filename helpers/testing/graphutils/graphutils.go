package graphutils

import (
	"errors"

	"github.com/asteris-llc/converge/graph"
)

// DependsOn returns true if target and source both exist in the graph, and
// there is a direct or transative dependency on target from source.
func DependsOn(g *graph.Graph, source, target string) bool {
	for _, dep := range g.Dependencies(source) {
		if dep == target {
			return true
		}
	}
	return false
}

// AddMetadata will insert metadata into the graph at the provided node.  It
// returns an error if the id doesn't exist or the metadata key does.
func AddMetadata(g *graph.Graph, id, key string, value interface{}) error {
	meta, ok := g.Get(id)
	if !ok {
		return errors.New("no such element: " + id)
	}
	return meta.AddMetadata(key, value)
}

// GetMetadata will lookup the metadata key from a graph node and return it.  It
// returns an error if the node doesn't exist, or the metadata key isn't found.
func GetMetadata(g *graph.Graph, id, key string) (interface{}, error) {
	meta, ok := g.Get(id)
	if !ok {
		return nil, errors.New("no such element: " + id)
	}
	value, found := meta.LookupMetadata(key)
	if !found {
		return nil, errors.New("no metadata value found: " + key)
	}
	return value, nil
}
