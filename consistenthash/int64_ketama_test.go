package consistenthash

import (
	"slices"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInt64Ring(t *testing.T) {
	t.Parallel()

	t.Run("default virtual spots", func(t *testing.T) {
		t.Parallel()

		r := NewInt64Ring(0)

		if r.virtualSpots != DefaultVirtualSpots {
			t.Errorf("Expected %d virtual spots, got %d", DefaultVirtualSpots, r.virtualSpots)
		}
	})

	t.Run("custom virtual spots", func(t *testing.T) {
		t.Parallel()

		const customSpots = 200
		r := NewInt64Ring(customSpots)

		if r.virtualSpots != customSpots {
			t.Errorf("Expected %d virtual spots, got %d", customSpots, r.virtualSpots)
		}
	})
}

func TestInt64HashRing_AddNode(t *testing.T) {
	t.Parallel()

	r := NewInt64Ring(100)
	nodes := []string{"node1", "node2", "node3"}

	err := r.AddNode(nodes[0])
	require.Nil(t, err)

	if len(r.nodes) != 100 {
		t.Errorf("Expected 100 virtual nodes, got %d", len(r.nodes))
	}

	err = r.AddNode(nodes[1])
	require.Nil(t, err)
	err = r.AddNode(nodes[2])
	require.Nil(t, err)

	if r.Len() != 300 {
		t.Errorf("Expected 300 virtual nodes, got %d", r.Len())
	}

	originalCount := r.Len()
	err = r.AddNode(nodes[0])
	require.Nil(t, err)

	if r.Len() != originalCount+100 {
		t.Errorf("Expected %d virtual nodes after duplicate add, got %d", originalCount+100, r.Len())
	}
}

func TestInt64HashRing_RemoveNode(t *testing.T) {
	t.Parallel()

	r := NewInt64Ring(50)
	nodes := []string{"nodeA", "nodeB", "nodeC"}

	for _, n := range nodes {
		err := r.AddNode(n)
		require.Nil(t, err)
	}

	//nolint:paralleltest
	t.Run("remove existing node", func(t *testing.T) {
		r.RemoveNode(nodes[1])

		for _, n := range r.nodes {
			assert.NotEqual(t, n.nodeName, nodes[1])
		}
	})

	//nolint:paralleltest
	t.Run("remove non-existent node", func(t *testing.T) {
		originalCount := r.Len()
		r.RemoveNode("ghost_node")
		assert.Equal(t, r.Len(), originalCount)
	})

	//nolint:paralleltest
	t.Run("remove all nodes", func(t *testing.T) {
		for _, n := range nodes {
			r.RemoveNode(n)
		}

		assert.Equal(t, r.Len(), 0)
	})
}

func TestInt64HashRing_GetNode(t *testing.T) {
	t.Parallel()

	r := NewInt64Ring(100)
	nodes := []string{"server1", "server2", "server3"}

	for _, n := range nodes {
		err := r.AddNode(n)
		require.Nil(t, err)
	}

	// Test key distribution
	distribution := make(map[string]int)

	const testKeys = 10_000

	for i := range testKeys {
		key := int64(i)
		node, ok := r.GetNode(key)

		if ok {
			distribution[node]++
		}
	}

	for node, count := range distribution {
		t.Logf("Node %s received %d keys (%.1f%%)", node, count, float64(count)/testKeys*100)
	}

	t.Run("empty ring", func(t *testing.T) {
		t.Parallel()

		emptyRing := NewInt64Ring(100)
		_, ok := emptyRing.GetNode(123)
		assert.False(t, ok)
	})

	t.Run("consistent hashing", func(t *testing.T) {
		t.Parallel()

		key := int64(42)
		node1, ok := r.GetNode(key)
		require.True(t, ok)
		node2, ok := r.GetNode(key)
		require.True(t, ok)
		assert.Equal(t, node1, node2)
	})

	t.Run("ring wrap-around", func(t *testing.T) {
		t.Parallel()
		// Find the highest hash value
		maxHash := r.nodes[r.Len()-1].hash
		testKey := maxHash + 1                                     // Force wrap-around
		node, ok := r.GetNode(int64(testKey & 0x7FFFFFFFFFFFFFFF)) //nolint:gosec // Acceptable for tests
		require.True(t, ok)
		assert.True(t, slices.Contains(nodes, node))
	})
}

func TestInt64HashRing_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	r := NewInt64Ring(160)

	var wg sync.WaitGroup

	// Concurrent writers
	for i := range 4 {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			for j := 0; j < 100; j++ {
				err := r.AddNode("node" + strconv.Itoa(id*100+j))
				require.Nil(t, err)
				r.RemoveNode("node" + strconv.Itoa((id*100+j)-1))
			}
		}(i)
	}

	// Concurrent readers
	for range 4 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for j := range 1000 {
				_, _ = r.GetNode(int64(j))
			}
		}()
	}

	wg.Wait()
}

func BenchmarkInt64HashRing_GetNode(b *testing.B) {
	r := NewInt64Ring(160)
	for i := range 10 {
		err := r.AddNode("node" + strconv.Itoa(i))
		require.Nil(b, err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := r.GetNode(int64(b.N))
			require.Nil(b, err)
		}
	})
}

func BenchmarkHashRing_AddNode(b *testing.B) {
	r := NewInt64Ring(160)

	nodeNames := make([]string, b.N)
	for i := range nodeNames {
		nodeNames[i] = "node" + strconv.Itoa(i)
	}

	b.ResetTimer()

	for i := range b.N {
		err := r.AddNode(nodeNames[i])
		require.Nil(b, err)
	}
}

func BenchmarkInt64HashRing_RemoveNode(b *testing.B) {
	r := NewInt64Ring(160)
	for i := range 10 {
		err := r.AddNode("node" + strconv.Itoa(i))
		require.Nil(b, err)
	}

	b.ResetTimer()

	for i := range b.N {
		r.RemoveNode("node" + strconv.Itoa(i))
	}
}
