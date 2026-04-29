package utils

type Stack[T any] []T

func (s *Stack[T]) Push(v T) {
	*s = append(*s, v)
}

func (s *Stack[T]) Pop() (v T) {
	if s.IsEmpty() {
		return
	}

	res := s.Peek()

	if s.Len() == 1 {
		*s = []T{}
	} else {
		*s = (*s)[:len(*s)-1]
	}

	return res
}

func (s Stack[T]) Peek() (v T) {
	if s.IsEmpty() {
		return
	}

	return s[len(s)-1]
}

func (s Stack[T]) IsEmpty() bool {
	return s.Len() == 0
}

func (s Stack[T]) Len() int {
	return len(s)
}

func (s Stack[T]) Copy() []T {
	return append(make([]T, 0), s...)
}

type Queue[T any] Stack[T]

func (q *Queue[T]) Push(v T) {
	*q = append(*q, v)
}

func (q *Queue[T]) Pop() (v T) {
	if q.IsEmpty() {
		return
	}

	res := q.Peek()

	if q.Len() == 1 {
		*q = []T{}
	} else {
		*q = (*q)[1:]
	}

	return res
}

func (q Queue[T]) Peek() (v T) {
	if q.IsEmpty() {
		return
	}

	return q[0]
}

func (q Queue[T]) IsEmpty() bool {
	return q.Len() == 0
}

func (q Queue[T]) Len() int {
	return len(q)
}

func (q Queue[T]) Raw() []T {
	return append(make([]T, 0), q...)
}
