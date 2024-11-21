package fastats

import (
	"testing"
	"fmt"
)

func exampleBedGraph(y func(BedEntry[float64]) bool) {
	b := BedEntry[float64]{ChrSpan: ChrSpan{Chr: "chr1", Span: Span{Start: 5, End: 25}}, Fields: 1.18}
	if !y(b) {
		return
	}
	b = BedEntry[float64]{ChrSpan: ChrSpan{Chr: "chr1", Span: Span{Start: 25, End: 30}}, Fields: 1.19}
	if !y(b) {
		return
	}
}

func TestSpreadBed(t *testing.T) {
	s := SpreadBed(exampleBedGraph)
	i := 0
	for b := range s {
		fmt.Printf("spread %v %v\n", i, b)
		i++
	}
}

func TestWindowBed(t *testing.T) {
	w := WindowBed(exampleBedGraph, 5, 2)
	i := 0
	for b := range w {
		fmt.Printf("windowed %v %v\n", i, b)
		i++
	}
}

func TestWindowBedWeak(t *testing.T) {
	w := WindowBedWeak(exampleBedGraph, 5, 2)
	i := 0
	for b := range w {
		fmt.Printf("windowedWeak %v %v\n", i, b)
		i++
	}
}

func TestAutoCorrelation(t *testing.T) {
	a := AutoCorrelationWindows(exampleBedGraph, 1, 5, 2)
	i := 0
	for b := range a {
		fmt.Printf("autocorrelation %v %v\n", i, b)
		i++
	}
}
