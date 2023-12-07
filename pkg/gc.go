package fastats

import (
	"io"
	"fmt"
	"flag"
	"bufio"
	"os"
	"github.com/jgbaldwinbrown/iter"
)

func GCFrac(seq string) float64 {
	gcCount := 0.0
	bpCount := 0.0
	for _, c := range []byte(seq) {
		switch c {
		case 'g': fallthrough
		case 'G': fallthrough
		case 'c': fallthrough
		case 'C':
			gcCount++
			fallthrough
		case 'a': fallthrough
		case 'A': fallthrough
		case 't': fallthrough
		case 'T':
			bpCount++
		case 'n':
		case 'N':
		default:
		}
	}
	return gcCount / bpCount
}

func GCIter(views iter.Iter[BedEntry[string]]) *iter.Iterator[BedEntry[float64]] {
	return &iter.Iterator[BedEntry[float64]]{Iteratef: func(yield func(BedEntry[float64]) error) error {
		return views.Iterate(func(view BedEntry[string]) error {
			frac := GCFrac(view.Fields)
			return yield(BedEntry[float64]{ChrSpan: view.ChrSpan, Fields: frac})
		})
	}}
}

func WriteGC(w io.Writer, it iter.Iter[BedEntry[float64]]) (n int, err error) {
	err = it.Iterate(func(b BedEntry[float64]) error {
		nwritten, e := fmt.Fprintf(w, "%v\t%v\t%v\t%v\n", b.Chr, b.Start, b.End, b.Fields)
		n += nwritten
		return e
	})
	return n, err
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
