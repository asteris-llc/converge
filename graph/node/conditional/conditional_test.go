package conditional_test

import (
	"errors"
	"testing"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/graph/node"
	"github.com/asteris-llc/converge/graph/node/conditional"
	"github.com/asteris-llc/converge/helpers/testing/graphutils"
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

func peersToIDs(in []*node.Node) (out []string) {
	out = make([]string, 0)
	for _, n := range in {
		out = append(out, n.ID)
	}
	return
}

func addPredicateMetadata(g *graph.Graph, id, switchName, caseName, predicate string, peers []string) {
	graphutils.AddMetadata(g, id, conditional.MetaSwitchName, switchName)
	graphutils.AddMetadata(g, id, conditional.MetaBranchName, caseName)
	graphutils.AddMetadata(g, id, conditional.MetaUnrenderedPredicate, predicate)
	graphutils.AddMetadata(g, id, conditional.MetaPeers, peers)
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

	addPredicateMetadata(g, "root/b", "switch-b", "case1", "true", []string{"b"})
	addPredicateMetadata(g, "root/d/a", "switch-d", "case2", "true", []string{"a", "b", "c"})
	addPredicateMetadata(g, "root/d/b", "switch-d", "case2", "true", []string{"a", "b", "c"})
	addPredicateMetadata(g, "root/d/c", "switch-d", "case2", "true", []string{"a", "b", "c"})

	return g
}
