package flow

import "sync"

// Data is a small concurrency-safe key/value bag used for flow-level shared
// state and for step-local scratch space between tasks.
type Data struct {
	mu sync.RWMutex
	m  map[string]any
}

func NewData() *Data { return &Data{m: map[string]any{}} }

func (d *Data) Set(k string, v any) {
	d.mu.Lock()
	d.m[k] = v
	d.mu.Unlock()
}

func (d *Data) Get(k string) (any, bool) {
	d.mu.RLock()
	v, ok := d.m[k]
	d.mu.RUnlock()
	return v, ok
}

// Get is a typed getter: v, ok := flow.Get[string](data, "user").
func Get[T any](d *Data, k string) (T, bool) {
	var zero T
	raw, ok := d.Get(k)
	if !ok {
		return zero, false
	}
	v, ok := raw.(T)
	return v, ok
}

// Scratch keys used by the built-in task pipeline.
const (
	ScratchRequest     = "request"
	ScratchResponse    = "response"
	ScratchToolResults = "tool_results"
	ScratchResult      = "result"
)
