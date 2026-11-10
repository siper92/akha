package flow_v3

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type graph struct {
	order []StepName
	edges map[StepName]map[Transition]nexStep
}

func parseGraph(def string) (graph, error) {
	var doc struct {
		Graph yaml.Node `yaml:"graph"`
	}

	if err := yaml.Unmarshal([]byte(strings.TrimSpace(def)), &doc); err != nil {
		return graph{}, err
	}

	if doc.Graph.Kind != yaml.MappingNode {
		return graph{}, fmt.Errorf("%w: 'graph' must be a mapping of steps", ErrInvalidGraph)
	}

	g := graph{edges: make(map[StepName]map[Transition]nexStep)}
	seen := make(map[StepName]bool)
	add := func(name StepName) {
		if !seen[name] && name != FinishStep {
			seen[name] = true
			g.order = append(g.order, name)
		}
	}

	nodes := doc.Graph.Content
	for i := 0; i+1 < len(nodes); i += 2 {
		name := StepName(nodes[i].Value)

		var edges map[Transition]nexStep
		if err := nodes[i+1].Decode(&edges); err != nil {
			return graph{}, fmt.Errorf("step %s: %w", name, err)
		}

		add(name)
		g.edges[name] = edges
	}

	for i := 0; i < len(g.order); i++ {
		for _, e := range g.edges[g.order[i]] {
			add(e.Next)
		}
	}

	return g, nil
}

func stepDefToTransitionMap(def string) (map[StepName]map[Transition]nexStep, error) {
	g, err := parseGraph(def)
	if err != nil {
		return nil, err
	}

	return g.edges, nil
}

func NewMachineFlow[D any](def string, execs map[StepName]StepExec[D]) (*MachineFlow[D], error) {
	g, err := parseGraph(def)
	if err != nil {
		return nil, err
	}

	steps := make([]Step[DataServiceProvide[D]], 0, len(g.order))
	for _, name := range g.order {
		exec, ok := execs[name]
		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrNoStepExec, name)
		}

		steps = append(steps, &MachineStep[D]{
			name:        name,
			transitions: g.edges[name],
			exec:        exec,
		})
	}

	m := &MachineFlow[D]{}
	if err := m.Init(steps...); err != nil {
		return nil, err
	}

	return m, nil
}
