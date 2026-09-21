package tu

import (
	"errors"
	"reflect"
	"testing"
)

type T struct{ *testing.T }

func New(t *testing.T) *T {
	return &T{T: t}
}

func Run[I, E any](t *T, cases []Case[I, E], fn func(I) (E, error), eq func(want, got E) bool) {
	t.Helper()
	if eq == nil {
		eq = func(want, got E) bool { return reflect.DeepEqual(want, got) }
	}
	for _, c := range cases {
		t.Run(c.Name, func(st *testing.T) {
			st.Helper()
			got, err := fn(c.Input)
			if c.Err != nil {
				if !errors.Is(err, c.Err) {
					st.Errorf("%s: want error %v, got %v", c.Name, c.Err, err)
				}
				return
			}
			if err != nil {
				st.Errorf("%s: unexpected error: %v", c.Name, err)
				return
			}
			if !eq(c.Expected, got) {
				st.Errorf("%s: want %#v, got %#v", c.Name, c.Expected, got)
			}
		})
	}
}
