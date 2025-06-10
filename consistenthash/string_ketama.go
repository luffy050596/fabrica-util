package consistenthash

import (
	"encoding/binary"
	"hash"
	"sort"
	"strconv"
	"sync"

	"github.com/go-pantheon/fabrica-util/errors"
	"github.com/spaolacci/murmur3"
)

const (
	// DefaultVirtualSpots is the default number of virtual spots per node (160)
	DefaultVirtualSpots = 160
)

type ringNode struct {
	nodeName string
	key      string
	hash     uint32
}

type ringNodes []ringNode

func (r ringNodes) Len() int           { return len(r) }
func (r ringNodes) Less(i, j int) bool { return r[i].hash < r[j].hash }
func (r ringNodes) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }

// HashRing implements a string-based consistent hash ring
type HashRing struct {
	mu sync.RWMutex

	virtualSpots int
	nodes        ringNodes
	hashCache    sync.Pool
}

// NewRing creates a new string-based consistent hash ring with the specified number of virtual spots
func NewRing(virtualSpots int) *HashRing {
	if virtualSpots <= 0 {
		virtualSpots = DefaultVirtualSpots
	}

	return &HashRing{
		virtualSpots: virtualSpots,
		hashCache: sync.Pool{
			New: func() interface{} {
				return murmur3.New64()
			},
		},
	}
}

// AddNode add node and sort automatically
func (h *HashRing) AddNode(nodeName string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	hash := h.hashCache.Get().(hash.Hash)
	defer h.hashCache.Put(hash)

	nodes := make(ringNodes, 0, h.virtualSpots)

	for i := range h.virtualSpots {
		key := nodeName + ":" + strconv.Itoa(i)

		hash.Reset()

		_, err := hash.Write([]byte(key))
		if err != nil {
			return errors.Wrap(err, "write to hasher failed")
		}

		hashBytes := hash.Sum(nil)

		// use binary package to read uint32 more efficiently
		nodes = append(nodes, ringNode{
			nodeName: nodeName,
			key:      key,
			hash:     binary.BigEndian.Uint32(hashBytes[len(hashBytes)-4:]),
		})
	}

	h.nodes = append(h.nodes, nodes...)
	sort.Sort(h.nodes)

	return nil
}

// RemoveNode removes a node with the given name from the hash ring
func (h *HashRing) RemoveNode(nodeName string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	filtered := h.nodes[:0]

	for _, n := range h.nodes {
		if n.nodeName != nodeName {
			filtered = append(filtered, n)
		}
	}

	h.nodes = filtered
}

// GetNode returns the node name for the given key
// It uses consistent hashing to find the appropriate node
func (h *HashRing) GetNode(key string) (string, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if len(h.nodes) == 0 {
		return "", false
	}

	hash := h.hashCache.Get().(hash.Hash)
	defer h.hashCache.Put(hash)

	hash.Reset()
	hash.Write([]byte(key))
	hashBytes := hash.Sum(nil)
	targetHash := binary.BigEndian.Uint32(hashBytes[len(hashBytes)-4:])

	idx := sort.Search(len(h.nodes), func(i int) bool {
		return h.nodes[i].hash >= targetHash
	})

	if idx == len(h.nodes) {
		idx = 0
	}

	return h.nodes[idx].nodeName, true
}

func (h *HashRing) Len() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return len(h.nodes)
}
