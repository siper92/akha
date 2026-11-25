package flow

import (
	"context"
	"time"
)

var _ FlowMachine[FlowService] = (*BasicFlow)(nil)

type BasicFlow struct {
	runID     string
	steps     []FlowStep[FlowService, DataStore]
	stepGraph StepGraph
	startStep StepID
	name      string
}

func NewBasicFlow(startStep StepID, name string) *BasicFlow {
	return &BasicFlow{
		startStep: startStep,
		name:      name,
		runID:     time.Now().Format("20060102150405"),
	}
}

func (b *BasicFlow) Init(steps []FlowStep[FlowService, DataStore], stepGraph StepGraph) {
	b.runID = time.Now().Format("20060102150405")
	b.steps = steps
	b.stepGraph = stepGraph
}

func (b *BasicFlow) Run(ctx context.Context) FlowResult {
	//TODO implement me
	panic("implement me")
}

func (b *BasicFlow) Info() RunInfo {
	//TODO implement me
	panic("implement me")
}

func (b *BasicFlow) Name() string {
	return "basic_flow"
}
