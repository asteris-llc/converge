package hclutils

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/load"
)

// LoadFromString loads an HCL file from a string returning Graph parse.Node
func LoadFromString(name, src string) (*graph.Graph, error) {
	if !strings.HasSuffix(name, ".hcl") {
		name = name + ".hcl"
	}
	tmpdir, err := ioutil.TempDir("", "converge-testing")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpdir)
	fileName := filepath.Join(tmpdir, name)
	ioutil.WriteFile(fileName, []byte(src), 0777)
	return load.Nodes(context.Background(), fileName, false)
}

// LoadAndParseFromString loads a file and parses it, returning
// Graph resource.Resource
func LoadAndParseFromString(name, src string) (*graph.Graph, error) {
	if !strings.HasSuffix(name, ".hcl") {
		name = name + ".hcl"
	}
	tmpdir, err := ioutil.TempDir("", "converge-testing")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpdir)
	fileName := filepath.Join(tmpdir, name)
	ioutil.WriteFile(fileName, []byte(src), 0777)
	return load.Load(context.Background(), fileName, false)
}
