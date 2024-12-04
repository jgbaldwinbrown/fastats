package fastats

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"iter"
	"os"
)

func GCFrac(seq string) float64 {
	gcCount := 0.0
	bpCount := 0.0
	for _, c := range []byte(seq) {
		switch c {
		case 'g':
			fallthrough
		case 'G':
			fallthrough
		case 'c':
			fallthrough
		case 'C':
			gcCount++
			fallthrough
		case 'a':
			fallthrough
		case 'A':
			fallthrough
		case 't':
			fallthrough
		case 'T':
			bpCount++
		case 'n':
		case 'N':
		default:
		}
	}
	return gcCount / bpCount
}

func GCIter[B BedEnter[string]](views iter.Seq2[B, error]) iter.Seq2[BedEntry[float64], error] {
	return func(yield func(BedEntry[float64], error) bool) {
		for view, e := range views {
			frac := GCFrac(view.BedFields())
			if !yield(BedEntry[float64]{ChrSpan: toChrSpan(view), Fields: frac}, e) {
				return
			}
		}
	}
}

func WriteGC[B BedEnter[float64]](w io.Writer, it iter.Seq2[B, error]) (n int, err error) {
	for b, e := range it {
		if e != nil {
			return n, e
		}
		nwritten, e := fmt.Fprintf(w, "%v\t%v\t%v\t%v\n", b.SpanChr(), b.SpanStart(), b.SpanEnd(), b.BedFields())
		n += nwritten
		if e != nil {
			return n, e
		}
	}
	return n, nil
}

func RunGC() {
	sizep := flag.Int("size", 1, "Window size")
	stepp := flag.Int("step", 1, "Window step distance")
	flag.Parse()

	r := bufio.NewReader(os.Stdin)
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	fait := ParseFasta(r)
	wins := FaWins(fait, int64(*sizep), int64(*stepp))
	gc := GCIter(wins)

	_, e := WriteGC(w, gc)
	if e != nil {
		panic(e)
	}
}
