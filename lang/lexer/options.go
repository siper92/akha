package lexer

type Option func(*options)

type options struct {
	comments bool
}

func WithComments() Option {
	return func(o *options) {
		o.comments = true
	}
}
