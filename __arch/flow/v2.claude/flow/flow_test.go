package flow

import (
	"context"
	"errors"
	"testing"
	"time"
)

func okStep(name string) Step {
	return StepFunc(name, func(context.Context, *StepContext) (Result, error) {
		return Result{Output: name}, nil
	})
}

func TestLinearFlowCompletes(t *testing.T) {
	def, err := New("lin").Then(okStep("a"), okStep("b")).Build()
	if err != nil {
		t.Fatal(err)
	}
	inst, err := def.Run(context.Background(), NewRegistry())
	if err != nil || inst.State() != FlowCompleted || len(inst.History()) != 2 {
		t.Fatalf("state=%s err=%v hist=%d", inst.State(), err, len(inst.History()))
	}
}

func TestRetryThenSucceed(t *testing.T) {
	calls := 0
	s := StepFunc("flaky", func(context.Context, *StepContext) (Result, error) {
		calls++
		if calls < 3 {
			return Result{}, errors.New("boom")
		}
		return Result{Output: "ok"}, nil
	}, WithRetry(2, 0))
	def, _ := New("r").Then(s).Build()
	inst, err := def.Run(context.Background(), NewRegistry())
	if err != nil || inst.History()[0].Attempts != 3 {
		t.Fatalf("err=%v attempts=%d", err, inst.History()[0].Attempts)
	}
}

func TestFailAndContinuePolicies(t *testing.T) {
	bad := StepFunc("bad", func(context.Context, *StepContext) (Result, error) {
		return Result{}, errors.New("nope")
	})
	def, _ := New("f").Then(bad, okStep("after")).Build()
	inst, err := def.Run(context.Background(), NewRegistry())
	if err == nil || inst.State() != FlowFailed {
		t.Fatalf("want failed, got %s", inst.State())
	}

	cont := StepFunc("bad", bad.fn, WithOnError(ContinueFlow))
	def2, _ := New("c").Then(cont, okStep("after")).Build()
	inst2, err := def2.Run(context.Background(), NewRegistry())
	if err != nil || inst2.State() != FlowCompleted {
		t.Fatalf("want completed, got %s %v", inst2.State(), err)
	}
}

func TestLoopGuardAndRouting(t *testing.T) {
	loop := StepFunc("loop", func(context.Context, *StepContext) (Result, error) {
		return Result{Next: "loop"}, nil
	})
	def, _ := New("l").Then(loop).MaxSteps(5).Build()
	inst, err := def.Run(context.Background(), NewRegistry())
	if err == nil || inst.State() != FlowFailed || len(inst.History()) != 5 {
		t.Fatalf("state=%s hist=%d", inst.State(), len(inst.History()))
	}
}

func TestTimeoutAndCancel(t *testing.T) {
	slow := StepFunc("slow", func(ctx context.Context, _ *StepContext) (Result, error) {
		<-ctx.Done()
		return Result{}, ctx.Err()
	}, WithTimeout(10*time.Millisecond))
	def, _ := New("t").Then(slow).Build()
	inst, _ := def.Run(context.Background(), NewRegistry())
	if inst.State() != FlowFailed {
		t.Fatalf("timeout should fail flow, got %s", inst.State())
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	inst2, _ := def.Run(ctx, NewRegistry())
	if inst2.State() != FlowCancelled {
		t.Fatalf("want cancelled, got %s", inst2.State())
	}
}

type stubAPI struct{ text string }

func (s stubAPI) Call(context.Context, Request) (Response, error) {
	return Response{Text: s.text, ToolCalls: []ToolCall{{Name: "echo", Args: map[string]any{"v": 1}}}}, nil
}

func TestAIStepUsesTaskProviderOverrides(t *testing.T) {
	sp := NewRegistry().
		Register(KeyAPI, APIClient(stubAPI{"flow"})).
		Register(KeyTools, NewToolbox(ToolFunc{"echo", func(_ context.Context, a map[string]any) (any, error) { return a["v"], nil }}))

	base := AIStep("ask", AIStepSpec{
		Build: func(*TaskProvider) (Request, error) { return Request{Prompt: "hi"}, nil },
		Tools: true,
	})
	over := base.Renamed("ask2").With(WithServices(func(r *Registry) {
		r.Register(KeyAPI, APIClient(stubAPI{"step"}))
	}))

	def, _ := New("ai").Then(base, over).Build()
	inst, err := def.Run(context.Background(), sp)
	if err != nil {
		t.Fatal(err)
	}
	o1, _ := inst.Output("ask")
	o2, _ := inst.Output("ask2")
	if o1.(AIOutput).Text != "flow" || o2.(AIOutput).Text != "step" {
		t.Fatalf("got %v / %v", o1, o2)
	}
	if o1.(AIOutput).ToolResults[0].Output != 1 {
		t.Fatalf("tool not executed: %+v", o1)
	}
}

func TestIllegalTransition(t *testing.T) {
	def, _ := New("x").Then(okStep("a")).Build()
	inst, _ := def.Run(context.Background(), NewRegistry())
	if err := inst.Run(context.Background()); err == nil {
		t.Fatal("re-running a completed flow must fail")
	}
}
