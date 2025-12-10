package fastats

type OrComment[T any] struct {
	Value T
	Comment string
	IsComment bool
}
