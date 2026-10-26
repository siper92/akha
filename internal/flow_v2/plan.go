package flow_v2

import (
	"fmt"
	"strings"
)

var _ ExecutionPlan = (*Graph)(nil)

type Graph struct {
	start string
	edges map[string]map[Outcome]string
}

func NewGraph(start string) *Graph {
	return &Graph{start: start, edges: map[string]map[Outcome]string{}}
}

func (g *Graph) Edge(from string, outcome Outcome, to string) *Graph {
	if g.edges[from] == nil {
		g.edges[from] = map[Outcome]string{}
	}
	g.edges[from][outcome] = to
	return g
}

func ParseGraph(text string) (*Graph, error) {
	var g *Graph
	for n, line := range strings.Split(text, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		id, rest, ok := strings.Cut(line, "-")
		if !ok {
			return nil, fmt.Errorf("%w: line %d", ErrBadPlan, n+1)
		}
		id = strings.TrimSpace(id)
		if g == nil {
			g = NewGraph(id)
		}
		for _, e := range strings.Split(rest, "|") {
			outcome, to, ok := strings.Cut(e, ":")
			if !ok {
				return nil, fmt.Errorf("%w: line %d", ErrBadPlan, n+1)
			}
			g.Edge(id, Outcome(strings.TrimSpace(outcome)), strings.TrimSpace(to))
		}
	}
	if g == nil {
		return nil, fmt.Errorf("%w: empty", ErrBadPlan)
	}
	return g, nil
}

func (g *Graph) Start() string { return g.start }

func (g *Graph) Next(id string, outcome Outcome) (string, error) {
	to, ok := g.edges[id][outcome]
	if !ok {
		return "", fmt.Errorf("%w: %s -%s->", ErrNoTransition, id, outcome)
	}
	return to, nil
}
