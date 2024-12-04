package fastats

import (
	"fmt"
	"testing"
)

func testingBed(y func(ChrSpan) bool) {
	vals := []ChrSpan {
		ChrSpan{Chr: "one", Span: Span{Start: 5, End: 8}},
		ChrSpan{Chr: "one", Span: Span{Start: 6, End: 15}},
		ChrSpan{Chr: "two", Span: Span{Start: 5, End: 8}},
	}
	for _, val := range vals {
		if !y(val) {
			return
		}
	}
}

func TestBedWinImplicitChrs(t *testing.T) {
	wins := BedWinImplicitChrs(testingBed, 3, 2)
	for win, vals := range wins {
		fmt.Printf("%#v; %#v\n", win, vals)
	}
}
