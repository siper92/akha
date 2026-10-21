package types

type FlowStep interface {
	ID() string
	Process(data DataUnit) (DataUnit, error)
}

type Flow interface {
	Load(src string) error
	Process(step string) error // process step by UD
	Exec() error               // process steps in order defined for the flow
}

type GenerateUnit interface {
	DataUnit
}

type GenerateFlow interface {
	Flow
	transform(data DataUnit) (DataUnit, error)
	check() error
	rate() (int, error)
	regen() error
	run() (out DataUnit, err error)
}
