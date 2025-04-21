package consistenthash

import (
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetInfo(t *testing.T) {
	t.Parallel()

	ring := NewRing(16)

	nodes := []string{
		"test.server.com#1",
		"test.server.com#2",
		"test.server.com#3",
		"test.server.com#4",
	}

	for _, k := range nodes {
		err := ring.AddNode(k)
		require.Nil(t, err)
	}

	m := make(map[string]int)

	for i := 0; i < 1e6; i++ {
		node, ok := ring.GetNode("test value" + strconv.FormatUint(uint64(i&0x7FFFFFFFFFFFFFFF), 10)) //nolint:gosec // acceptable for tests
		assert.True(t, ok)

		m[node]++
	}

	for i := 1; i < len(nodes); i++ {
		ring.RemoveNode(nodes[i])
	}

	m = make(map[string]int)

	for i := 0; i < 1e6; i++ {
		node, ok := ring.GetNode("test value" + strconv.FormatUint(uint64(i&0x7FFFFFFFFFFFFFFF), 10)) //nolint:gosec // acceptable for tests
		assert.True(t, ok)

		m[node]++
	}

	ring.RemoveNode(nodes[0])

	for i := 0; i < 1e6; i++ {
		node, ok := ring.GetNode("test value" + strconv.FormatUint(uint64(i&0x7FFFFFFFFFFFFFFF), 10)) //nolint:gosec // acceptable for tests
		assert.False(t, ok)
		assert.Equal(t, node, "")
	}
}

func TestNewRing(t *testing.T) {
	t.Parallel()

	t.Run("default virtual spots", func(t *testing.T) {
		t.Parallel()

		r := NewRing(0)
		assert.Equal(t, r.virtualSpots, DefaultVirtualSpots)
	})

	t.Run("custom virtual spots", func(t *testing.T) {
		t.Parallel()

		customSpots := 200
		r := NewRing(customSpots)
		assert.Equal(t, r.virtualSpots, customSpots)
	})
}

func TestHashRing_AddRemoveNodes(t *testing.T) {
	t.Parallel()

	r := NewRing(100)
	nodes := []string{"node1", "node2", "node3"}

	t.Run("add nodes", func(t *testing.T) {
		t.Parallel()

		for _, n := range nodes {
			err := r.AddNode(n)
			require.Nil(t, err)
		}

		assert.Equal(t, len(r.nodes), len(nodes)*r.virtualSpots)
	})

	t.Run("remove node", func(t *testing.T) {
		t.Parallel()
		r.RemoveNode("node2")
		expected := (len(nodes) - 1) * r.virtualSpots
		assert.Equal(t, len(r.nodes), expected)

		for _, n := range r.nodes {
			assert.NotEqual(t, n.nodeName, "node2")
		}
	})
}

func TestHashRing_GetNode(t *testing.T) {
	t.Parallel()

	r := NewRing(100)

	nodes := []string{"nodeA", "nodeB", "nodeC"}
	for _, n := range nodes {
		err := r.AddNode(n)
		require.Nil(t, err)
	}

	testCases := []struct {
		key      string
		expected string
	}{
		{"user123", ""},
		{"session-abc", ""},
		{"data:1", ""},
		{"config:prod", ""},
	}

	// First pass to record distribution
	distribution := make(map[string]string)

	for _, tc := range testCases {
		node, found := r.GetNode(tc.key)
		assert.True(t, found)

		distribution[tc.key] = node
	}

	t.Run("consistent distribution", func(t *testing.T) {
		t.Parallel()

		for _, tc := range testCases {
			node, _ := r.GetNode(tc.key)
			assert.Equal(t, node, distribution[tc.key])
		}
	})

	t.Run("empty ring", func(t *testing.T) {
		t.Parallel()

		emptyRing := NewRing(100)
		node, found := emptyRing.GetNode("anykey")
		assert.False(t, found)
		assert.Equal(t, node, "")
	})

	t.Run("wrap around", func(t *testing.T) {
		t.Parallel()
		// Create predictable ring with known hash values
		r := NewRing(1)
		err := r.AddNode("nodeX")
		require.Nil(t, err)
		err = r.AddNode("nodeY")
		require.Nil(t, err)

		// Force wrap around scenario
		highHashKey := "zzzzzzzzzzzzzzzz"
		node, _ := r.GetNode(highHashKey)
		assert.Equal(t, node, r.nodes[1].nodeName)
	})
}

func TestHashRing_Consistency(t *testing.T) {
	t.Parallel()

	r := NewRing(100)

	initialNodes := []string{"node1", "node2", "node3"}
	for _, n := range initialNodes {
		err := r.AddNode(n)
		require.Nil(t, err)
	}

	keys := make([]string, 1000)
	original := make(map[string]string)

	for i := range keys {
		keys[i] = "key" + strconv.Itoa(i)
		original[keys[i]], _ = r.GetNode(keys[i])
	}

	t.Run("after adding node", func(t *testing.T) {
		t.Parallel()

		err := r.AddNode("node4")
		require.Nil(t, err)

		changed := 0

		for _, k := range keys {
			node, _ := r.GetNode(k)
			if node != original[k] {
				changed++
			}
		}

		t.Logf("Changed keys after adding node: %.2f%%", float64(changed)/float64(len(keys))*100)
	})

	t.Run("after removing node", func(t *testing.T) {
		t.Parallel()
		r.RemoveNode("node3")

		changed := 0

		for _, k := range keys {
			node, _ := r.GetNode(k)
			if node != original[k] {
				changed++
			}
		}

		t.Logf("Changed keys after removing node: %.2f%%", float64(changed)/float64(len(keys))*100)
	})
}

func TestConcurrentAccess(t *testing.T) {
	t.Parallel()

	r := NewRing(100)

	var wg sync.WaitGroup

	wg.Add(3)

	go func() {
		defer wg.Done()

		for i := 0; i < 100; i++ {
			err := r.AddNode("node" + strconv.Itoa(i))
			require.Nil(t, err)
		}
	}()

	go func() {
		defer wg.Done()

		for i := 0; i < 100; i++ {
			r.RemoveNode("node" + strconv.Itoa(i))
		}
	}()

	go func() {
		defer wg.Done()

		for i := 0; i < 1000; i++ {
			r.GetNode("key" + strconv.Itoa(i))
		}
	}()

	wg.Wait() // Should not panic
}

func BenchmarkHashRing_GetNode(b *testing.B) {
	r := NewRing(100)
	for i := 0; i < 10; i++ {
		err := r.AddNode("node" + strconv.Itoa(i))
		require.Nil(b, err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r.GetNode("some-key")
		}
	})
}

func BenchmarkAddNode(b *testing.B) {
	r := NewRing(200)

	b.ResetTimer()

	for i := range b.N {
		err := r.AddNode("node" + strconv.Itoa(i))
		require.Nil(b, err)
	}
}

func BenchmarkRemoveNode(b *testing.B) {
	r := NewRing(100)
	for i := 0; i < 10; i++ {
		err := r.AddNode("node" + strconv.Itoa(i))
		require.Nil(b, err)
	}

	b.ResetTimer()

	for i := range b.N {
		r.RemoveNode("node" + strconv.Itoa(i%100))
	}
}
