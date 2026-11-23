package flow

import (
	"context"
	"fmt"
	"sync"
)

// ---------------------------------------------------------------------------
// Flow service provider (SP)
//
// The flow owns a Provider: a registry of services (API clients, tools,
// loggers, config...). Steps never construct services themselves; they
// receive them from the flow, which keeps steps reusable across flows.
// ---------------------------------------------------------------------------

// Provider resolves services by key.
type Provider interface {
	Lookup(key string) (any, bool)
}

// Registry is the default Provider. Registries can be chained: a child
// looks up locally first, then falls back to its parent. This is how a
// step/task scopes or overrides services without touching the flow's SP.
type Registry struct {
	mu       sync.RWMutex
	services map[string]any
	parent   Provider
}

func NewRegistry() *Registry { return &Registry{services: map[string]any{}} }

// Child returns a new registry that falls back to r.
func (r *Registry) Child() *Registry {
	c := NewRegistry()
	c.parent = r
	return c
}

// Register adds or replaces a service. Returns r for chaining.
func (r *Registry) Register(key string, svc any) *Registry {
	r.mu.Lock()
	r.services[key] = svc
	r.mu.Unlock()
	return r
}

func (r *Registry) Lookup(key string) (any, bool) {
	r.mu.RLock()
	svc, ok := r.services[key]
	r.mu.RUnlock()
	if ok {
		return svc, true
	}
	if r.parent != nil {
		return r.parent.Lookup(key)
	}
	return nil, false
}

// Resolve fetches a typed service from any Provider.
func Resolve[T any](p Provider, key string) (T, error) {
	var zero T
	if p == nil {
		return zero, fmt.Errorf("provider: nil provider looking up %q", key)
	}
	raw, ok := p.Lookup(key)
	if !ok {
		return zero, fmt.Errorf("provider: service %q not registered", key)
	}
	svc, ok := raw.(T)
	if !ok {
		return zero, fmt.Errorf("provider: service %q has type %T, want %T", key, raw, zero)
	}
	return svc, nil
}

// ---------------------------------------------------------------------------
// Well-known service contracts. Flows plug in concrete implementations
// (OpenAI, Anthropic, internal HTTP, mocks...), steps only see interfaces.
// ---------------------------------------------------------------------------

const (
	KeyAPI   = "api"
	KeyTools = "tools"
)

// Request / Response are a deliberately generic API envelope.
type Request struct {
	Prompt string
	Tools  []string       // names of tools the callee may request
	Meta   map[string]any // provider-specific options
}

type ToolCall struct {
	Name string
	Args map[string]any
}

type Response struct {
	Text      string
	ToolCalls []ToolCall
	Raw       any
}

// APIClient is any remote call: an LLM, a REST service, etc.
type APIClient interface {
	Call(ctx context.Context, req Request) (Response, error)
}

// Tool is something a task can execute locally on behalf of the API.
type Tool interface {
	Name() string
	Run(ctx context.Context, args map[string]any) (any, error)
}

// ToolFunc adapts a function into a Tool.
type ToolFunc struct {
	ToolName string
	Fn       func(ctx context.Context, args map[string]any) (any, error)
}

func (t ToolFunc) Name() string { return t.ToolName }
func (t ToolFunc) Run(ctx context.Context, args map[string]any) (any, error) {
	return t.Fn(ctx, args)
}

// Toolbox is a named set of tools, registered under KeyTools.
type Toolbox map[string]Tool

func NewToolbox(tools ...Tool) Toolbox {
	tb := Toolbox{}
	for _, t := range tools {
		tb[t.Name()] = t
	}
	return tb
}
