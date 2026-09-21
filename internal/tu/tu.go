package tu

type Case[I, E any] struct {
	Name     string
	Input    I
	Expected E
	Err      error
}
