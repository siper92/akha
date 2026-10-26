package flow

import (
	"context"

	"github.com/siper92/akha/internal/data_sources"
	"github.com/siper92/akha/internal/state_machine"
)

type flowInput struct {
	nums    []int
	workers int
	source  string
}

func calcTask[D any](id ID, fn func(d D, s Scope[int]) int) *MachineTask[D, int, string] {
	m := state_machine.NewMachine[TaskState[D, int], string]("calc").
		Step("calc", func(_ context.Context, st *TaskState[D, int]) (string, error) {
			st.Result = fn(st.Data, st.Scope)
			return "done", nil
		}, "done").
		Final("done")
	return NewTask[D, int, string](id, m)
}

func newFlow(in flowInput) (*Runner, error) {
	reg := data_sources.NewRegistry()
	err := reg.Register(data_sources.NewSource("nums", data_sources.KindOther, func(context.Context, data_sources.Query) ([]int, error) {
		return in.nums, nil
	}))
	if err != nil {
		return nil, err
	}

	double, err := NewStep[int, int]("double", FromSource[int](in.source, data_sources.Query{}),
		calcTask("x2", func(d int, _ Scope[int]) int { return d * 2 }),
		calcTask("inc", func(_ int, s Scope[int]) int {
			r, _ := s.Result("x2")
			return r + 1
		}),
	)
	if err != nil {
		return nil, err
	}
	double.Workers(in.workers)

	sum, err := NewStep[[]int, int]("sum",
		DataGetterFunc[[]int](func(_ context.Context, s StepScope) ([][]int, error) {
			v, err := ResultAs[[]int](s, "double")
			return [][]int{v}, err
		}),
		calcTask("total", func(d []int, _ Scope[int]) int {
			t := 0
			for _, v := range d {
				t += v
			}
			return t
		}),
	)
	if err != nil {
		return nil, err
	}

	return New(NewPlan("double", "sum"), reg, ServiceMap{}, double, sum)
}
