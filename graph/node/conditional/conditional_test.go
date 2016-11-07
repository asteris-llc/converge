package conditional_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/graph/node"
	"github.com/asteris-llc/converge/graph/node/conditional"
	"github.com/asteris-llc/converge/helpers/testing/graphutils"
	"github.com/asteris-llc/converge/render/extensions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestIsConditional(t *testing.T) {
	g := sampleGraph()

	t.Run("when-non-conditional-node", func(t *testing.T) {
		meta, ok := g.Get("root/a")
		require.True(t, ok)
		assert.False(t, conditional.IsConditional(meta))
	})
	t.Run("when-single-conditional-node", func(t *testing.T) {
		meta, ok := g.Get("root/b")
		require.True(t, ok)
		assert.True(t, conditional.IsConditional(meta))
	})
	t.Run("when-first", func(t *testing.T) {
		meta, ok := g.Get("root/d/a")
		require.True(t, ok)
		assert.True(t, conditional.IsConditional(meta))
	})
	t.Run("when-middle", func(t *testing.T) {
		meta, ok := g.Get("root/d/b")
		require.True(t, ok)
		assert.True(t, conditional.IsConditional(meta))
	})
	t.Run("when-last", func(t *testing.T) {
		meta, ok := g.Get("root/d/c")
		require.True(t, ok)
		assert.True(t, conditional.IsConditional(meta))
	})
}

func TestPeerNodes(t *testing.T) {
	g := sampleGraph()

	t.Run("when-non-conditional-node", func(t *testing.T) {
		expected := []string{}
		meta, ok := g.Get("root/a")
		require.True(t, ok)
		assert.Equal(t, expected, peersToIDs(conditional.PeerNodes(g, meta)))
	})
	t.Run("when-single-conditional-node", func(t *testing.T) {
		expected := []string{"root/b"}
		meta, ok := g.Get("root/b")
		require.True(t, ok)
		assert.Equal(t, expected, peersToIDs(conditional.PeerNodes(g, meta)))
	})
	t.Run("when-many-conditional-nodes", func(t *testing.T) {
		expected := []string{"root/d/a", "root/d/b", "root/d/c"}
		for _, id := range expected {
			t.Run(id, func(t *testing.T) {
				meta, ok := g.Get(id)
				require.True(t, ok)
				assert.Equal(t, expected, peersToIDs(conditional.PeerNodes(g, meta)))
			})
		}
	})
	t.Run("when-missing", func(t *testing.T) {
		g := sampleGraph()
		meta, _ := g.Get("root/c")
		graphutils.AddMetadata(g, "root/c", conditional.MetaPeers, []string{"root/missing-node"})
		assert.Nil(t, conditional.PeerNodes(g, meta))
	})
}

func TestPeerBranches(t *testing.T) {
	g := peerBranchSampleGraph(t)
	t.Run("when-non-conditional-node", func(t *testing.T) {
		meta, _ := g.Get("root/resource1")
		assert.Equal(t, []string{}, peersToIDs(conditional.PeerBranches(g, meta)))
	})
	t.Run("when-resource-node", func(t *testing.T) {
		switches := []string{"root/switch1", "root/switch2", "root/switch3"}
		cases := []string{"case1", "case2", "case3"}
		resources := []string{"resource1", "resource2", "resource3"}
		for _, switchID := range switches {
			casePeers := []string{}
			for _, caseID := range cases {
				casePeers = append(casePeers, fmt.Sprintf("%s/%s", switchID, caseID))
			}
			for _, caseID := range cases {
				for _, resourceID := range resources {
					node := fmt.Sprintf("%s/%s/%s", switchID, caseID, resourceID)
					meta, ok := g.Get(node)
					require.True(t, ok)
					assert.Equal(t, casePeers, peersToIDs(conditional.PeerBranches(g, meta)))
				}
			}
		}
	})

	t.Run("when-case-node", func(t *testing.T) {
		switches := []string{"root/switch1", "root/switch2", "root/switch3"}
		cases := []string{"case1", "case2", "case3"}
		for _, switchID := range switches {
			casePeers := []string{}
			for _, caseID := range cases {
				casePeers = append(casePeers, fmt.Sprintf("%s/%s", switchID, caseID))
			}
			for _, caseID := range cases {
				node := fmt.Sprintf("%s/%s", switchID, caseID)
				meta, ok := g.Get(node)
				require.True(t, ok)
				assert.Equal(t, casePeers, peersToIDs(conditional.PeerBranches(g, meta)))
			}
		}
	})

	t.Run("when-switch-node", func(t *testing.T) {
		meta, _ := g.Get("root/switch1")
		assert.Equal(t, []string{}, peersToIDs(conditional.PeerBranches(g, meta)))
	})
}

func TestRenderPredicate(t *testing.T) {
	t.Run("when-not-rendered", func(t *testing.T) {
		g := sampleGraph()
		renderer := NewRenderer("id1", "result1", nil)
		meta, _ := g.Get("root/b")
		_, ok := meta.LookupMetadata(conditional.MetaRenderedPredicate)
		require.False(t, ok)
		result, err := conditional.RenderPredicate(meta, renderer.Render)
		assert.NoError(t, err)
		assert.Equal(t, "result1", result)
		renderer.AssertCalled(t, "Render", mock.Anything, mock.Anything)
	})
	t.Run("when-previously-rendered", func(t *testing.T) {
		g := sampleGraph()
		renderer := NewRenderer("id1", "result1", nil)
		meta, _ := g.Get("root/b")
		graphutils.AddMetadata(g, "root/b", conditional.MetaRenderedPredicate, "pre-rendered")
		result, err := conditional.RenderPredicate(meta, renderer.Render)
		assert.NoError(t, err)
		assert.Equal(t, "pre-rendered", result)
		renderer.AssertNotCalled(t, "Render", mock.Anything, mock.Anything)
	})
	t.Run("when-render-error", func(t *testing.T) {
		expectedErr := errors.New("i am error")
		g := sampleGraph()
		renderer := NewRenderer("", "", expectedErr)
		meta, _ := g.Get("root/b")
		_, err := conditional.RenderPredicate(meta, renderer.Render)
		assert.Equal(t, expectedErr, err)
	})
}

func TestIsTrue(t *testing.T) {
	t.Run("errors-when-unrendered", func(t *testing.T) {
		g := sampleGraph()
		meta, _ := g.Get("root/b")
		_, err := conditional.IsTrue(meta)
		assert.Error(t, err)
	})
	t.Run("returns-true-when-truthy", func(t *testing.T) {
		truthyValues := []string{"t", "T", "true", "True", "TRUE", "trUe"}
		for _, truth := range truthyValues {
			g := sampleGraph()
			meta, _ := g.Get("root/b")
			graphutils.AddMetadata(g, "root/b", conditional.MetaRenderedPredicate, truth)
			ok, err := conditional.IsTrue(meta)
			assert.NoError(t, err)
			assert.True(t, ok)
		}
	})
	t.Run("returns-false-when-untruthy", func(t *testing.T) {
		untruthyValues := []string{"f", "false", "False", "FALSE", "0", "trooo", ""}
		for _, truth := range untruthyValues {
			g := sampleGraph()
			meta, _ := g.Get("root/b")
			graphutils.AddMetadata(g, "root/b", conditional.MetaRenderedPredicate, truth)
			ok, err := conditional.IsTrue(meta)
			assert.NoError(t, err)
			assert.False(t, ok)
		}
	})
	t.Run("returns-cached-value-when-untruthy", func(t *testing.T) {
		untruthyValues := []string{"f", "false", "False", "FALSE", "0", "trooo", ""}
		for _, truth := range untruthyValues {
			g := sampleGraph()
			meta, _ := g.Get("root/b")
			graphutils.AddMetadata(g, "root/b", conditional.MetaRenderedPredicate, truth)
			graphutils.AddMetadata(g, "root/b", conditional.MetaPredicate, true)
			ok, err := conditional.IsTrue(meta)
			assert.NoError(t, err)
			assert.True(t, ok)
		}
	})
	t.Run("returns-cached-value-when-truthy", func(t *testing.T) {
		truthyValues := []string{"t", "T", "true", "True", "TRUE", "trUe"}
		for _, truth := range truthyValues {
			g := sampleGraph()
			meta, _ := g.Get("root/b")
			graphutils.AddMetadata(g, "root/b", conditional.MetaRenderedPredicate, truth)
			graphutils.AddMetadata(g, "root/b", conditional.MetaPredicate, false)
			ok, err := conditional.IsTrue(meta)
			assert.NoError(t, err)
			assert.False(t, ok)
		}
	})
}

func TestShouldEvaluate(t *testing.T) {

	t.Run("when-many-branches", func(t *testing.T) {
		g := peerBranchSampleGraph(t)
		resources := []string{"resource1", "resource2", "resource3"}
		t.Run("when-true-true-true", func(t *testing.T) {
			t.Run("first-branch", func(t *testing.T) {
				branch := "root/switch1/case1"
				for _, res := range resources {
					id := graph.ID(branch, res)
					meta, ok := g.Get(id)
					require.True(t, ok)
					shouldEval, err := conditional.ShouldEvaluate(g, meta)
					require.NoError(t, err)
					assert.True(t, shouldEval)
				}
			})
			t.Run("second-branch", func(t *testing.T) {
				branch := "root/switch1/case2"
				for _, res := range resources {
					id := graph.ID(branch, res)
					meta, ok := g.Get(id)
					require.True(t, ok)
					shouldEval, err := conditional.ShouldEvaluate(g, meta)
					require.NoError(t, err)
					assert.False(t, shouldEval)
				}
			})
			t.Run("third-branch", func(t *testing.T) {
				branch := "root/switch1/case3"
				for _, res := range resources {
					id := graph.ID(branch, res)
					meta, ok := g.Get(id)
					require.True(t, ok)
					shouldEval, err := conditional.ShouldEvaluate(g, meta)
					require.NoError(t, err)
					assert.False(t, shouldEval)
				}
			})
		})
		t.Run("when-false-true-true", func(t *testing.T) {
			t.Run("first-branch", func(t *testing.T) {
				branch := "root/switch2/case1"
				for _, res := range resources {
					id := graph.ID(branch, res)
					meta, ok := g.Get(id)
					require.True(t, ok)
					shouldEval, err := conditional.ShouldEvaluate(g, meta)
					require.NoError(t, err)
					assert.False(t, shouldEval)
				}
			})
			t.Run("second-branch", func(t *testing.T) {
				branch := "root/switch2/case2"
				for _, res := range resources {
					id := graph.ID(branch, res)
					meta, ok := g.Get(id)
					require.True(t, ok)
					shouldEval, err := conditional.ShouldEvaluate(g, meta)
					require.NoError(t, err)
					assert.True(t, shouldEval)
				}
			})
			t.Run("third-branch", func(t *testing.T) {
				branch := "root/switch2/case3"
				for _, res := range resources {
					id := graph.ID(branch, res)
					meta, ok := g.Get(id)
					require.True(t, ok)
					shouldEval, err := conditional.ShouldEvaluate(g, meta)
					require.NoError(t, err)
					assert.False(t, shouldEval)
				}
			})
		})
		t.Run("when-false-false-true", func(t *testing.T) {
			t.Run("first-branch", func(t *testing.T) {
				branch := "root/switch3/case1"
				for _, res := range resources {
					id := graph.ID(branch, res)
					meta, ok := g.Get(id)
					require.True(t, ok)
					shouldEval, err := conditional.ShouldEvaluate(g, meta)
					require.NoError(t, err)
					assert.False(t, shouldEval)
				}
			})
			t.Run("second-branch", func(t *testing.T) {
				branch := "root/switch3/case2"
				for _, res := range resources {
					id := graph.ID(branch, res)
					meta, ok := g.Get(id)
					require.True(t, ok)
					shouldEval, err := conditional.ShouldEvaluate(g, meta)
					require.NoError(t, err)
					assert.False(t, shouldEval)
				}
			})
			t.Run("third-branch", func(t *testing.T) {
				branch := "root/switch3/case3"
				for _, res := range resources {
					id := graph.ID(branch, res)
					meta, ok := g.Get(id)
					require.True(t, ok)
					shouldEval, err := conditional.ShouldEvaluate(g, meta)
					require.NoError(t, err)
					assert.True(t, shouldEval)
				}
			})
		})
	})

	t.Run("when-error", func(t *testing.T) {
		g := sampleGraph()
		meta, _ := g.Get("root/a")
		_, err := conditional.ShouldEvaluate(g, meta)
		assert.Error(t, err)
	})
}

func peersToIDs(in []*node.Node) (out []string) {
	out = make([]string, 0)
	for _, n := range in {
		out = append(out, n.ID)
	}
	return
}

func addPredicateMetadata(g *graph.Graph, id, switchName, caseName, nodeType, predicate string, peers []string) {
	if switchName != "" {
		graphutils.AddMetadata(g, id, conditional.MetaSwitchName, switchName)
	}
	if caseName != "" {
		graphutils.AddMetadata(g, id, conditional.MetaBranchName, caseName)
	}
	if predicate != "" {
		graphutils.AddMetadata(g, id, conditional.MetaUnrenderedPredicate, predicate)
	}
	if len(peers) > 0 {
		graphutils.AddMetadata(g, id, conditional.MetaPeers, peers)
	}
	if nodeType != "" {
		graphutils.AddMetadata(g, id, conditional.MetaType, nodeType)
	}
}

type MockRenderer struct {
	mock.Mock
}

func (m *MockRenderer) GetID() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockRenderer) Render(name, content string) (string, error) {
	args := m.Called(name, content)
	return args.String(0), args.Error(1)
}

func NewRenderer(id, renderVal string, err error) *MockRenderer {
	m := &MockRenderer{}
	m.On("GetID").Return(id)
	m.On("Render", mock.Anything, mock.Anything).Return(renderVal, err)
	return m
}

func sampleGraph() *graph.Graph {
	g := graph.New()

	g.Add(node.New(graph.ID("root"), struct{}{}))
	g.Add(node.New(graph.ID("root", "a"), struct{}{}))
	g.ConnectParent(graph.ID("root"), graph.ID("root", "a"))

	g.Add(node.New(graph.ID("root", "b"), struct{}{}))
	g.ConnectParent(graph.ID("root"), graph.ID("root", "b"))

	g.Add(node.New(graph.ID("root", "c"), struct{}{}))
	g.ConnectParent(graph.ID("root"), graph.ID("root", "c"))

	g.Add(node.New(graph.ID("root", "d"), struct{}{}))
	g.ConnectParent(graph.ID("root"), graph.ID("root", "d"))

	g.Add(node.New(graph.ID("root", "d", "a"), struct{}{}))
	g.ConnectParent(graph.ID("root", "d"), graph.ID("root", "d", "a"))

	g.Add(node.New(graph.ID("root", "d", "b"), struct{}{}))
	g.ConnectParent(graph.ID("root", "d"), graph.ID("root", "d", "b"))

	g.Add(node.New(graph.ID("root", "d", "c"), struct{}{}))
	g.ConnectParent(graph.ID("root", "d"), graph.ID("root", "d", "c"))

	g.Add(node.New(graph.ID("root", "d", "c", "a"), struct{}{}))
	g.ConnectParent(graph.ID("root", "d", "c"), graph.ID("root", "d", "c", "a"))

	g.Add(node.New(graph.ID("root", "d", "c", "b"), struct{}{}))
	g.ConnectParent(graph.ID("root", "d", "c"), graph.ID("root", "d", "c", "b"))

	addPredicateMetadata(g, "root/b", "switch-b", "case1", conditional.NodeCatBranch, "true", []string{"b"})
	addPredicateMetadata(g, "root/d/a", "switch-d", "case2", conditional.NodeCatBranch, "true", []string{"a", "b", "c"})
	addPredicateMetadata(g, "root/d/b", "switch-d", "case2", conditional.NodeCatBranch, "true", []string{"a", "b", "c"})
	addPredicateMetadata(g, "root/d/c", "switch-d", "case2", conditional.NodeCatBranch, "true", []string{"a", "b", "c"})

	return g
}

func peerBranchSampleGraph(t *testing.T) *graph.Graph {
	g := graph.New()

	g.Add(node.New(graph.ID("root"), struct{}{}))
	g.Add(node.New(graph.ID("root", "switch1"), struct{}{}))
	g.Add(node.New(graph.ID("root", "switch2"), struct{}{}))
	g.Add(node.New(graph.ID("root", "switch3"), struct{}{}))

	g.Add(node.New(graph.ID("root", "resource1"), struct{}{}))
	g.Add(node.New(graph.ID("root", "resource2"), struct{}{}))
	g.Add(node.New(graph.ID("root", "resource3"), struct{}{}))

	g.ConnectParent("root", "root/switch1")
	g.ConnectParent("root", "root/switch2")
	g.ConnectParent("root", "root/switch3")

	g.ConnectParent("root", "root/resource1")
	g.ConnectParent("root", "root/resource2")
	g.ConnectParent("root", "root/resource3")

	g.Add(node.New(graph.ID("root", "switch1", "case1"), struct{}{}))
	g.Add(node.New(graph.ID("root", "switch1", "case1", "resource1"), struct{}{}))
	g.Add(node.New(graph.ID("root", "switch1", "case1", "resource2"), struct{}{}))
	g.Add(node.New(graph.ID("root", "switch1", "case1", "resource3"), struct{}{}))

	g.Add(node.New(graph.ID("root", "switch1", "case2"), struct{}{}))
	g.Add(node.New(graph.ID("root", "switch1", "case2", "resource1"), struct{}{}))
	g.Add(node.New(graph.ID("root", "switch1", "case2", "resource2"), struct{}{}))
	g.Add(node.New(graph.ID("root", "switch1", "case2", "resource3"), struct{}{}))

	g.Add(node.New(graph.ID("root", "switch1", "case3"), struct{}{}))
	g.Add(node.New(graph.ID("root", "switch1", "case3", "resource1"), struct{}{}))
	g.Add(node.New(graph.ID("root", "switch1", "case3", "resource2"), struct{}{}))
	g.Add(node.New(graph.ID("root", "switch1", "case3", "resource3"), struct{}{}))

	g.ConnectParent("root/switch1", "root/switch1/case1")
	g.ConnectParent("root/switch1/case1", "root/switch1/case1/resource1")
	g.ConnectParent("root/switch1/case1", "root/switch1/case1/resource2")
	g.ConnectParent("root/switch1/case1", "root/switch1/case1/resource3")

	g.ConnectParent("root/switch1", "root/switch1/case2")
	g.ConnectParent("root/switch1/case2", "root/switch1/case2/resource1")
	g.ConnectParent("root/switch1/case2", "root/switch1/case2/resource2")
	g.ConnectParent("root/switch1/case2", "root/switch1/case2/resource3")

	g.ConnectParent("root/switch1", "root/switch1/case3")
	g.ConnectParent("root/switch1/case3", "root/switch1/case3/resource1")
	g.ConnectParent("root/switch1/case3", "root/switch1/case3/resource2")
	g.ConnectParent("root/switch1/case3", "root/switch1/case3/resource3")

	addPredicateMetadata(g, "root/switch1", "switch1", "", conditional.NodeCatSwitch, "", []string{})
	addPredicateMetadata(g, "root/switch1/case1", "switch1", "case1", conditional.NodeCatBranch, "true", []string{"case1", "case2", "case3"})
	addPredicateMetadata(g, "root/switch1/case2", "switch1", "case2", conditional.NodeCatBranch, "true", []string{"case1", "case2", "case3"})
	addPredicateMetadata(g, "root/switch1/case3", "switch1", "case3", conditional.NodeCatBranch, "true", []string{"case1", "case2", "case3"})

	addPredicateMetadata(g, "root/switch1/case1/resource1", "switch1", "case1", conditional.NodeCatResource, "true", []string{"case1", "case2", "case3"})
	addPredicateMetadata(g, "root/switch1/case1/resource2", "switch1", "case1", conditional.NodeCatResource, "true", []string{"case1", "case2", "case3"})
	addPredicateMetadata(g, "root/switch1/case1/resource3", "switch1", "case1", conditional.NodeCatResource, "true", []string{"case1", "case2", "case3"})

	addPredicateMetadata(g, "root/switch1/case2/resource1", "switch1", "case2", conditional.NodeCatResource, "true", []string{"case1", "case2", "case3"})
	addPredicateMetadata(g, "root/switch1/case2/resource2", "switch1", "case2", conditional.NodeCatResource, "true", []string{"case1", "case2", "case3"})
	addPredicateMetadata(g, "root/switch1/case2/resource3", "switch1", "case2", conditional.NodeCatResource, "true", []string{"case1", "case2", "case3"})

	addPredicateMetadata(g, "root/switch1/case3/resource1", "switch1", "case3", conditional.NodeCatResource, "true", []string{"case1", "case2", "case3"})
	addPredicateMetadata(g, "root/switch1/case3/resource2", "switch1", "case3", conditional.NodeCatResource, "true", []string{"case1", "case2", "case3"})
	addPredicateMetadata(g, "root/switch1/case3/resource3", "switch1", "case3", conditional.NodeCatResource, "true", []string{"case1", "case2", "case3"})

	g.Add(node.New(graph.ID("root", "switch2", "case1"), struct{}{}))
	g.Add(node.New(graph.ID("root", "switch2", "case1", "resource1"), struct{}{}))
	g.Add(node.New(graph.ID("root", "switch2", "case1", "resource2"), struct{}{}))
	g.Add(node.New(graph.ID("root", "switch2", "case1", "resource3"), struct{}{}))

	g.Add(node.New(graph.ID("root", "switch2", "case2"), struct{}{}))
	g.Add(node.New(graph.ID("root", "switch2", "case2", "resource1"), struct{}{}))
	g.Add(node.New(graph.ID("root", "switch2", "case2", "resource2"), struct{}{}))
	g.Add(node.New(graph.ID("root", "switch2", "case2", "resource3"), struct{}{}))

	g.Add(node.New(graph.ID("root", "switch2", "case3"), struct{}{}))
	g.Add(node.New(graph.ID("root", "switch2", "case3", "resource1"), struct{}{}))
	g.Add(node.New(graph.ID("root", "switch2", "case3", "resource2"), struct{}{}))
	g.Add(node.New(graph.ID("root", "switch2", "case3", "resource3"), struct{}{}))

	g.ConnectParent("root/switch2", "root/switch2/case1")
	g.ConnectParent("root/switch2/case1", "root/switch2/case1/resource1")
	g.ConnectParent("root/switch2/case1", "root/switch2/case1/resource2")
	g.ConnectParent("root/switch2/case1", "root/switch2/case1/resource3")

	g.ConnectParent("root/switch2", "root/switch2/case2")
	g.ConnectParent("root/switch2/case2", "root/switch2/case2/resource1")
	g.ConnectParent("root/switch2/case2", "root/switch2/case2/resource2")
	g.ConnectParent("root/switch2/case2", "root/switch2/case2/resource3")

	g.ConnectParent("root/switch2", "root/switch2/case3")
	g.ConnectParent("root/switch2/case3", "root/switch2/case3/resource1")
	g.ConnectParent("root/switch2/case3", "root/switch2/case3/resource2")
	g.ConnectParent("root/switch2/case3", "root/switch2/case3/resource3")

	addPredicateMetadata(g, "root/switch2", "switch2", "", conditional.NodeCatSwitch, "", []string{})
	addPredicateMetadata(g, "root/switch2/case1", "switch2", "case1", conditional.NodeCatBranch, "false", []string{"case1", "case2", "case3"})
	addPredicateMetadata(g, "root/switch2/case2", "switch2", "case2", conditional.NodeCatBranch, "true", []string{"case1", "case2", "case3"})
	addPredicateMetadata(g, "root/switch2/case3", "switch2", "case3", conditional.NodeCatBranch, "true", []string{"case1", "case2", "case3"})

	addPredicateMetadata(g, "root/switch2/case1/resource1", "switch2", "case1", conditional.NodeCatResource, "false", []string{"case1", "case2", "case3"})
	addPredicateMetadata(g, "root/switch2/case1/resource2", "switch2", "case1", conditional.NodeCatResource, "false", []string{"case1", "case2", "case3"})
	addPredicateMetadata(g, "root/switch2/case1/resource3", "switch2", "case1", conditional.NodeCatResource, "false", []string{"case1", "case2", "case3"})

	addPredicateMetadata(g, "root/switch2/case2/resource1", "switch2", "case2", conditional.NodeCatResource, "true", []string{"case1", "case2", "case3"})
	addPredicateMetadata(g, "root/switch2/case2/resource2", "switch2", "case2", conditional.NodeCatResource, "true", []string{"case1", "case2", "case3"})
	addPredicateMetadata(g, "root/switch2/case2/resource3", "switch2", "case2", conditional.NodeCatResource, "true", []string{"case1", "case2", "case3"})

	addPredicateMetadata(g, "root/switch2/case3/resource1", "switch2", "case3", conditional.NodeCatResource, "true", []string{"case1", "case2", "case3"})
	addPredicateMetadata(g, "root/switch2/case3/resource2", "switch2", "case3", conditional.NodeCatResource, "true", []string{"case1", "case2", "case3"})
	addPredicateMetadata(g, "root/switch2/case3/resource3", "switch2", "case3", conditional.NodeCatResource, "true", []string{"case1", "case2", "case3"})

	g.Add(node.New(graph.ID("root", "switch3", "case1"), struct{}{}))
	g.Add(node.New(graph.ID("root", "switch3", "case1", "resource1"), struct{}{}))
	g.Add(node.New(graph.ID("root", "switch3", "case1", "resource2"), struct{}{}))
	g.Add(node.New(graph.ID("root", "switch3", "case1", "resource3"), struct{}{}))

	g.Add(node.New(graph.ID("root", "switch3", "case2"), struct{}{}))
	g.Add(node.New(graph.ID("root", "switch3", "case2", "resource1"), struct{}{}))
	g.Add(node.New(graph.ID("root", "switch3", "case2", "resource2"), struct{}{}))
	g.Add(node.New(graph.ID("root", "switch3", "case2", "resource3"), struct{}{}))

	g.Add(node.New(graph.ID("root", "switch3", "case3"), struct{}{}))
	g.Add(node.New(graph.ID("root", "switch3", "case3", "resource1"), struct{}{}))
	g.Add(node.New(graph.ID("root", "switch3", "case3", "resource2"), struct{}{}))
	g.Add(node.New(graph.ID("root", "switch3", "case3", "resource3"), struct{}{}))

	g.ConnectParent("root/switch3", "root/switch3/case1")
	g.ConnectParent("root/switch3/case1", "root/switch3/case1/resource1")
	g.ConnectParent("root/switch3/case1", "root/switch3/case1/resource2")
	g.ConnectParent("root/switch3/case1", "root/switch3/case1/resource3")

	g.ConnectParent("root/switch3", "root/switch3/case2")
	g.ConnectParent("root/switch3/case2", "root/switch3/case2/resource1")
	g.ConnectParent("root/switch3/case2", "root/switch3/case2/resource2")
	g.ConnectParent("root/switch3/case2", "root/switch3/case2/resource3")

	g.ConnectParent("root/switch3", "root/switch3/case3")
	g.ConnectParent("root/switch3/case3", "root/switch3/case3/resource1")
	g.ConnectParent("root/switch3/case3", "root/switch3/case3/resource2")
	g.ConnectParent("root/switch3/case3", "root/switch3/case3/resource3")

	addPredicateMetadata(g, "root/switch3", "switch3", "", conditional.NodeCatSwitch, "", []string{})

	addPredicateMetadata(g, "root/switch3/case1", "switch2", "case1", conditional.NodeCatBranch, "false", []string{"case1", "case2", "case3"})
	addPredicateMetadata(g, "root/switch3/case2", "switch2", "case2", conditional.NodeCatBranch, "false", []string{"case1", "case2", "case3"})
	addPredicateMetadata(g, "root/switch3/case3", "switch2", "case3", conditional.NodeCatBranch, "true", []string{"case1", "case2", "case3"})

	addPredicateMetadata(g, "root/switch3/case1/resource1", "switch3", "case1", conditional.NodeCatResource, "false", []string{"case1", "case2", "case3"})
	addPredicateMetadata(g, "root/switch3/case1/resource2", "switch3", "case1", conditional.NodeCatResource, "false", []string{"case1", "case2", "case3"})
	addPredicateMetadata(g, "root/switch3/case1/resource3", "switch3", "case1", conditional.NodeCatResource, "false", []string{"case1", "case2", "case3"})

	addPredicateMetadata(g, "root/switch3/case2/resource1", "switch3", "case2", conditional.NodeCatResource, "false", []string{"case1", "case2", "case3"})
	addPredicateMetadata(g, "root/switch3/case2/resource2", "switch3", "case2", conditional.NodeCatResource, "false", []string{"case1", "case2", "case3"})
	addPredicateMetadata(g, "root/switch3/case2/resource3", "switch3", "case2", conditional.NodeCatResource, "false", []string{"case1", "case2", "case3"})

	addPredicateMetadata(g, "root/switch3/case3/resource1", "switch3", "case3", conditional.NodeCatResource, "true", []string{"case1", "case2", "case3"})
	addPredicateMetadata(g, "root/switch3/case3/resource2", "switch3", "case3", conditional.NodeCatResource, "true", []string{"case1", "case2", "case3"})
	addPredicateMetadata(g, "root/switch3/case3/resource3", "switch3", "case3", conditional.NodeCatResource, "true", []string{"case1", "case2", "case3"})

	idRender := func(_, s string) (string, error) {
		lang := extensions.DefaultLanguage()
		results, err := lang.Render(struct{}{}, "test", s)
		require.NoError(t, err)
		return results.String(), nil
	}

	for _, s := range []string{"switch1", "switch2", "switch3"} {
		for _, c := range []string{"case1", "case2", "case3"} {
			cID := graph.ID("root", s, c)
			cMeta, _ := g.Get(cID)
			_, err := conditional.RenderPredicate(cMeta, idRender)
			require.NoError(t, err)
			for _, r := range []string{"resource1", "resource2", "resource3"} {
				nid := graph.ID(cID, r)
				nMeta, _ := g.Get(nid)
				_, err := conditional.RenderPredicate(nMeta, idRender)
				require.NoError(t, err)
			}
		}
	}

	return g
}
