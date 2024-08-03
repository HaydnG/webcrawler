package history

import "sync"

type History[T any] struct {
	lookup map[string]T
	sync.RWMutex
}

func NewHistory[T any]() *History[T] {
	return &History[T]{
		lookup: make(map[string]T, 20),
	}
}

func (h *History[T]) Check(key string) (T, bool) {
	h.RLock()
	v, ok := h.lookup[key]
	h.RUnlock()
	return v, ok
}

func (h *History[T]) Add(k string, v T) {
	h.Lock()
	h.lookup[k] = v
	h.Unlock()
}

// GetHistory used to access the history, not threadsafe
func (h *History[T]) GetKeys() []string {
	keys := make([]string, 0, len(h.lookup))

	// Iterate over the map and collect the keys
	for key := range h.lookup {
		keys = append(keys, key)
	}

	return keys
}
