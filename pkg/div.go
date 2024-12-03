package fastats

import (
	"bufio"
	"os"
	"fmt"
	"iter"
	"errors"
	"log"

	"github.com/jgbaldwinbrown/iterh"
)

type Tuple2[T, U any] struct {
	V1 T
	V2 U
}

func ChrSpanEq[C1, C2 ChrSpanner](c1 C1, c2 C2) bool {
	return c1.SpanChr() == c2.SpanChr() && c1.SpanStart() == c2.SpanStart() && c1.SpanEnd() == c2.SpanEnd()
}

var ErrZipShort = errors.New("Zip: beds are not the same length")
var ErrZipNoMatch = errors.New("Zip: bed entry spans do not match")

func ZipMatchedSorted[B1 BedEnter[T1], B2 BedEnter[T2], T1, T2 any](it1 iter.Seq[B1], it2 iter.Seq[B2]) iter.Seq2[BedEntry[Tuple2[T1, T2]], error] {
	return func(y func(BedEntry[Tuple2[T1, T2]], error) bool) {
		p2, cancel := iter.Pull(it2)
		defer cancel()
		for b1 := range it1 {
			b2, ok := p2()
			out := BedEntry[Tuple2[T1, T2]]{}
			out.ChrSpan = ToChrSpan(b1)
			out.Fields.V1 = b1.BedFields()
			out.Fields.V2 = b2.BedFields()

			if !ok {
				if !y(out, ErrZipShort) {
					return
				}
			}
			if !ChrSpanEq(b1, b2) {
				if !y(out, fmt.Errorf("b1 %v b2 %v: %w", b1, b2, ErrZipNoMatch)) {
					return
				}
			}
			if !y(out, nil) {
				return
			}
		}
	}
}

type OkTuple2[T, U any] struct {
	Tuple2[T, U]
	Ok bool
}

func ZipMatches[B1 BedEnter[T1], B2 BedEnter[T2], T1, T2 any](it1 iter.Seq[B1], it2 iter.Seq[B2]) map[ChrSpan]OkTuple2[T1, T2] {
	m := map[ChrSpan]OkTuple2[T1, T2]{}
	for b := range it1 {
		tup := OkTuple2[T1, T2]{}
		tup.V1 = b.BedFields()
		m[ToChrSpan(b)] = tup
	}
	for b := range it2 {
		cs := ToChrSpan(b)
		tup, found := m[cs]
		tup.V2 = b.BedFields()
		tup.Ok = found
		m[cs] = tup
	}
	return m
}

func IterateZipMap[T1, T2 any](m map[ChrSpan]OkTuple2[T1, T2]) iter.Seq[BedEntry[Tuple2[T1, T2]]] {
	return func(yield func(BedEntry[Tuple2[T1, T2]]) bool) {
		for key, val := range m {
			b := BedEntry[Tuple2[T1, T2]]{}
			b.ChrSpan = key
			b.Fields = val.Tuple2
			if val.Ok {
				if !yield(b) {
					return
				}
			}
		}
	}
}

func DivBed[B BedEnter[Tuple2[float64, float64]]](it iter.Seq[B]) iter.Seq[BedEntry[float64]] {
	return func(y func(BedEntry[float64]) bool) {
		for b := range it {
			out := BedEntry[float64]{}
			out.ChrSpan = ToChrSpan(b)
			fields := b.BedFields()
			out.Fields = fields.V1 / fields.V2
			if !y(out) {
				return
			}
		}
	}
}

func FullDivCovs() {
	if len(os.Args) < 3 {
		fmt.Printf("usage: %v bed1.bed bed2.bed\n", os.Args[0])
		log.Fatal(fmt.Errorf("Not enough args: %v", os.Args))
	}
	cov1, errp1 := iterh.BreakWithError(iterh.PathIter(os.Args[1], ParseBedGraph))
	scov1 := SpreadBed(cov1)
	rpkm1, _ := RpkmAndTotal(scov1)

	cov2, errp2 := iterh.BreakWithError(iterh.PathIter(os.Args[2], ParseBedGraph))
	scov2 := SpreadBed(cov2)
	rpkm2, _ := RpkmAndTotal(scov2)

	zipped := ZipMatches(rpkm1, rpkm2)
	div := DivBed(IterateZipMap(zipped))

	w := bufio.NewWriter(os.Stdout)
	defer func() {
		e := w.Flush()
		if e != nil {
			log.Fatal(e)
		}
	}()

	for b := range div {
		_, e := fmt.Fprintf(w, "%v\t%v\t%v\t%v\n", b.Chr, b.Start, b.End, b.Fields)
		if e != nil {
			log.Fatal(e)
		}
	}
	if *errp1 != nil {
		log.Fatal(*errp1)
	}
	if *errp2 != nil {
		log.Fatal(*errp2)
	}
}
