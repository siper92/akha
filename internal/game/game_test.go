package game

import (
	"context"
	"testing"

	ds "github.com/siper92/akha/internal/data_sources"
	flow "github.com/siper92/akha/internal/flow_v2"
	tu "github.com/siper92/akha/lang/tests_utils"
)

// --- game parse flow from stories and csv to GamePop

func TestGameParseFlow(t *testing.T) {
	ctx := context.Background()
	expected, err := ds.Read[[]GamePop](ctx, ds.JSONHandler[[]GamePop]()("testdata/resultJSON.json"))
	if err != nil {
		t.Fatal(err)
	}

	cases := []tu.Case[string, []GamePop]{
		{
			Name:     "stories_and_csv_build_game_pops",
			Input:    "testdata/execPlanSchema.txt",
			Expected: expected,
		},
	}

	tu.Run(tu.New(t), cases, func(planPath string) ([]GamePop, error) {
		text, err := ds.Read[string](ctx, ds.NewTextFile(planPath))
		if err != nil {
			return nil, err
		}
		plan, err := flow.ParseGraph(text)
		if err != nil {
			return nil, err
		}
		steps, err := Steps(plan)
		if err != nil {
			return nil, err
		}
		m := flow.NewMachine(20)
		if err := m.Init(plan, steps, ds.NewFolderProvider("testdata", Handlers())); err != nil {
			return nil, err
		}
		if err := m.Run(ctx); err != nil {
			return nil, err
		}
		return last[[]GamePop]("buildJSON")(ctx, m)
	}, nil)
}
