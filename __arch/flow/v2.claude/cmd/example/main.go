package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"example.com/flowsm/flow"
)

// ---------------------------------------------------------------------------
// Fake services (in real life: an LLM client, HTTP client, DB, ...)
// ---------------------------------------------------------------------------

type fakeLLM struct{ name string }

func (f fakeLLM) Call(_ context.Context, req flow.Request) (flow.Response, error) {
	resp := flow.Response{Text: fmt.Sprintf("[%s] %s", f.name, req.Prompt)}
	if strings.Contains(req.Prompt, "weather") {
		resp.ToolCalls = []flow.ToolCall{{Name: "weather", Args: map[string]any{"city": "Sofia"}}}
	}
	return resp, nil
}

var weatherTool = flow.ToolFunc{ToolName: "weather", Fn: func(_ context.Context, a map[string]any) (any, error) {
	return fmt.Sprintf("sunny in %v", a["city"]), nil
}}

// ---------------------------------------------------------------------------
// Reusable step library — no flow knows these internals, no step knows flows.
// ---------------------------------------------------------------------------

// Ask is a reusable AI step: prompt comes from shared flow data.
func Ask(name, dataKey string) *flow.TaskStep {
	return flow.AIStep(name, flow.AIStepSpec{
		Build: func(tp *flow.TaskProvider) (flow.Request, error) {
			q, _ := flow.Get[string](tp.Shared, dataKey)
			return flow.Request{Prompt: q, Tools: []string{"weather"}}, nil
		},
		Tools: true,
	}, flow.WithRetry(2, 10*time.Millisecond), flow.WithTimeout(time.Second))
}

// Store copies the previous AI output into shared data.
var Store = flow.StepFunc("store", func(_ context.Context, sc *flow.StepContext) (flow.Result, error) {
	sc.Data.Set("stored", true)
	return flow.Result{Output: "ok"}, nil
})

// Gate routes based on data: an example of branching.
var Gate = flow.StepFunc("gate", func(_ context.Context, sc *flow.StepContext) (flow.Result, error) {
	if v, _ := flow.Get[bool](sc.Data, "vip"); v {
		return flow.Result{Next: "vip_answer"}, nil
	}
	return flow.Result{Next: "answer"}, nil
})

func main() {
	ctx := context.Background()
	logger := func(e flow.Event) {
		if e.Step == "" {
			fmt.Printf("  flow  %-10s %v\n", e.Flow, errOrEmpty(e.Err))
		} else {
			fmt.Printf("  step  %-12s %-9s attempt=%d %v\n", e.Step, e.StepSt, e.Attempt, errOrEmpty(e.Err))
		}
	}

	// Flow SP: shared by every step in the flow.
	sp := flow.NewRegistry().
		Register(flow.KeyAPI, flow.APIClient(fakeLLM{name: "default-model"})).
		Register(flow.KeyTools, flow.NewToolbox(weatherTool))

	// --- Flow 1: linear --------------------------------------------------
	f1, err := flow.New("assistant").Then(Ask("answer", "question"), Store).Build()
	must(err)

	data := flow.NewData()
	data.Set("question", "what's the weather?")
	fmt.Println("assistant flow:")
	inst, err := f1.Run(ctx, sp, flow.WithData(data), flow.WithObserver(logger))
	must(err)
	out, _ := inst.Output("answer")
	fmt.Printf("  => %+v\n\n", out)

	// --- Flow 2: same steps reused + branching + step-scoped SP override --
	vip := Ask("answer", "question").Renamed("vip_answer").With(
		flow.WithServices(func(r *flow.Registry) {
			r.Register(flow.KeyAPI, flow.APIClient(fakeLLM{name: "premium-model"}))
		}),
	)
	done := flow.StepFunc("done", func(context.Context, *flow.StepContext) (flow.Result, error) {
		return flow.Result{Stop: true}, nil
	})
	f2, err := flow.New("router").
		Then(Gate, Ask("answer", "question"), done, vip, Store).
		Build()
	must(err)

	data2 := flow.NewData()
	data2.Set("question", "hello")
	data2.Set("vip", true)
	fmt.Println("router flow (vip):")
	inst2, err := f2.Run(ctx, sp, flow.WithData(data2), flow.WithObserver(logger))
	must(err)
	out2, _ := inst2.Output("vip_answer")
	fmt.Printf("  => %+v  state=%s\n", out2, inst2.State())
}

func errOrEmpty(err error) any {
	if err == nil {
		return ""
	}
	return err
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
