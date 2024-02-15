package fastats

import (
	"fmt"
	"testing"
)

func printDeque[T any](d *Deque[T]) {
	l := d.Len()
	for i := 0; i < l; i++ {
		fmt.Printf("\t%v", d.Get(i))
	}
	fmt.Printf("\n")
}

func TestDeque(t *testing.T) {
	var d Deque[int]
	for i := 0; i < 100000; i++ {
		d.PushBack(i)
	}
	for i := 0; i < 99990; i++ {
		d.PopFront()
	}
	fmt.Println("printing deque")
	fmt.Printf("%#v\n", d)
	printDeque(&d)
	fmt.Println("printed deque")
}
