package tests_utils

type Case[I, E any] struct {
	Name     string
	Input    I
	Expected E
	Err      error
}
