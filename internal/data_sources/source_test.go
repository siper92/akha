package data_sources

import (
	"context"
	"testing"

	tu "github.com/siper92/akha/lang/tests_utils"
)

// --- typed fetch from provider

func TestFetch(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Register(NewSource("nums", KindOther, func(context.Context, Query) ([]int, error) {
		return []int{1, 2}, nil
	}))

	cases := []tu.Case[string, []int]{
		{
			Name:     "registered_source",
			Input:    "nums",
			Expected: []int{1, 2},
		},
		{
			Name:  "unknown_source",
			Input: "missing",
			Err:   ErrSourceNotFound,
		},
	}

	tu.Run(tu.New(t), cases, func(name string) ([]int, error) {
		return Fetch[[]int](context.Background(), reg, name, Query{})
	}, nil)

	if _, err := Get[string](reg, "nums"); err == nil {
		t.Fatalf("expected %v", ErrSourceType)
	}
}
