package tu

import (
	"errors"
	"fmt"
	"testing"
)

var errNegative = errors.New("negative")

func double(n int) (int, error) {
	if n < 0 {
		return 0, fmt.Errorf("double %d: %w", n, errNegative)
	}
	return n * 2, nil
}

func TestRun(t *testing.T) {
	cases := []Case[int, int]{
		// --- doubles
		{
			Name:     "zero_is_zero",
			Input:    0,
			Expected: 0,
		},
		{
			Name:     "positive_is_doubled",
			Input:    21,
			Expected: 42,
		},
		// --- errors
		{
			Name:  "negative_is_error",
			Input: -1,
			Err:   errNegative,
		},
	}
	Run(New(t), cases, double, nil)
}

func TestRunCustomEq(t *testing.T) {
	cases := []Case[int, int]{
		// --- custom equality
		{
			Name:     "same_sign_matches",
			Input:    3,
			Expected: 100,
		},
	}
	Run(New(t), cases, double, func(want, got int) bool { return (want > 0) == (got > 0) })
}

func TestRunSlices(t *testing.T) {
	cases := []Case[int, []int]{
		// --- deep equal
		{
			Name:     "nil_matches_nil",
			Input:    0,
			Expected: nil,
		},
		{
			Name:     "elements_match",
			Input:    2,
			Expected: []int{0, 1},
		},
	}

	Run(New(t), cases, func(n int) ([]int, error) {
		if n == 0 {
			return nil, nil
		}
		out := make([]int, n)
		for i := range out {
			out[i] = i
		}
		return out, nil
	}, nil)
}
