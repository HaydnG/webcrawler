package history

import "sync"

// History is a generic thread safe map storage implementation
type History[T any] struct {
	lookup map[string]T
	sync.RWMutex
}

// NewHistory initializes a new History set
func NewHistory[T any]() *History[T] {
	return &History[T]{
		lookup: make(map[string]T, 100),
	}
}

// Get checks for existence, and returns the found value
func (h *History[T]) Get(key string) (T, bool) {
	h.RLock()
	v, ok := h.lookup[key]
	h.RUnlock()
	return v, ok
}

// Add inserts a new value into the History
func (h *History[T]) Add(k string, v T) {
	h.Lock()
	h.lookup[k] = v
	h.Unlock()
}

// GetHistory takes a snapshot of all the keys in the map, returns a slice of keys
func (h *History[T]) GetKeys() []string {
	h.RLock()
	defer h.RUnlock()

	keys := make([]string, 0, len(h.lookup))

	// Iterate over the map and collect the keys
	for key := range h.lookup {
		keys = append(keys, key)
	}

	return keys
}
