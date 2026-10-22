package flow

type FlowStep interface {
	ID() string
	Process() error
}

type Flow interface {
	Init(steps ...FlowStep) error
	Process(step string) (string, error) // process step by ID, return next step or error
	Exec() error                         // process steps in order defined for the flow
}

type LoadDataStep interface {
	FlowStep
}

type TransformDataStep interface {
	FlowStep
}

type CheckStep interface {
	FlowStep
}

type EvaluateStep interface {
	FlowStep
}

type RegenerateStep interface {
	FlowStep
}

type RunStep interface {
	FlowStep
}
