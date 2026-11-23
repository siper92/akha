package flow_redo_v3

// ---------------------------------------------------------------------
// Contracts
// ---------------------------------------------------------------------

type StepID string
type Outcome string

const (
	OutcomeFinal = "final"
)

// ---------------------------------------------------------------------
// Data and Service Provider
// ---------------------------------------------------------------------

type DataGetter interface {
	Get(key string) any
}

type DataSetter interface {
	Set(key string, value any) (bool, error)
}

type Data interface {
	DataGetter
	DataSetter
}

type ServiceProvider interface {
	Service(name string) any
}

// ---------------------------------------------------------------------
// Machine
// ---------------------------------------------------------------------

type Guard func(sp ServiceProvider) bool

type Transition struct {
	To   StepID
	When Guard
}

type FlowTransitionMap map[StepID]map[Outcome]Transition

// ---------------------------------------------------------------------
// util funcs
// ---------------------------------------------------------------------

func TypedData[T any](d DataGetter, key string) (T, bool) {
	var zero T
	v := d.Get(key)
	if v == nil {
		return zero, false
	}
	t, ok := v.(T)
	if !ok {
		return zero, false
	}
	return t, true
}

func TypedService[T any](sp ServiceProvider, name string) (T, bool) {
	var zero T
	v := sp.Service(name)
	if v == nil {
		return zero, false
	}
	t, ok := v.(T)
	if !ok {
		return zero, false
	}
	return t, true
}
