package fastats

import (
	"fmt"
)

type Deque[T any] struct {
	data []T
	start int
	length int
}

func (d *Deque[T]) Grow() {
	newcap := len(d.data) * 2
	if newcap < 1 {
		newcap = 1
	}
	newdata := make([]T, 0, newcap)
	newdata = d.AppendToSlice(newdata[:0])
	newdata = newdata[:newcap]
	d.data = newdata
	d.start = 0
}

func (d *Deque[T]) starti() int {
	return d.start % len(d.data)
}

func (d *Deque[T]) endi() int {
	return (d.start + d.length) % len(d.data)
}

func (d *Deque[T]) Len() int {
	return d.length
}

func (d *Deque[T]) Get(i int) T {
	if d.length <= i {
		panic(fmt.Errorf("Deque.Get: i %v not a safe index; Deque.Len %v", i, d.Len()))
	}
	return d.data[(d.start + i) % len(d.data)]
}

func (d *Deque[T]) Set(i int, val T) bool {
	if d.length <= i {
		return false
	}
	d.data[(d.start + i) % len(d.data)] = val
	return true
}

func (d *Deque[T]) PushBack(vals ...T) {
	newlen := d.length + len(vals)
	for newlen > len(d.data) {
		d.Grow()
	}

	for _, v := range vals {
		d.data[d.endi()] = v
		d.length++
	}
}

func (d *Deque[T]) PopFront() (val T, ok bool) {
	if d.length < 1 {
		return val, false
	}
	val = d.data[d.starti()]
	d.start++
	d.length--
	return val, true
}

func (d *Deque[T]) PushFront(val T) {
	newlen := d.length + 1
	for newlen > len(d.data) {
		d.Grow()
	}

	d.start = (d.start + d.length - 1) % len(d.data)
}

func (d *Deque[T]) PopBack() (val T, ok bool) {
	if d.length < 1 {
		return val, false
	}
	val = d.Get(d.length - 1)
	d.length--
	return val, true
}

func (d *Deque[T]) AppendToSlice(s []T) []T {
	l := d.Len()
	for i := 0; i < l; i++ {
		s = append(s, d.Get(i))
	}
	return s
}
