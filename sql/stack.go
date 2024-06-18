package sql

type stack[T any] []T

func (s *stack[T]) push(e T) {
	*s = append(*s, e)
}

func (s *stack[T]) pop() T {
	l := len(*s)
	if l == 0 {
		var noop T
		return noop
	}

	e := (*s)[l-1]
	*s = (*s)[:l-1]

	return e
}
