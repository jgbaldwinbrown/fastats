package fastats

func Transform[T, U any](it Iter[T], f func(T) (U, error)) *Iterator[U] {
	return &Iterator[U]{Iteratef: func(yield func(U) error) error {
		return it.Iterate(func(val T) error {
			valu, e := f(val)
			if e != nil {
				return e
			}
			return yield(valu)
		})
	}}
}
