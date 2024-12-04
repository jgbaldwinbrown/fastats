package fastats

import (
	"iter"
)

func Wins(start, end, size, step int64) iter.Seq[Span] {
	return func(yield func(Span) bool) {
		var i int64
		for i = start; i < end; i += step {
			winend := i + size
			if winend > end {
				winend = end
			}
			if !yield(Span{i, winend}) {
				return
			}
		}
	}
}

func FaEntryWins[F FaEnter](fe F, size int64, step int64) iter.Seq[BedEntry[string]] {
	return func(yield func(BedEntry[string]) bool) {
		wins := Wins(0, int64(len(fe.FaSeq())), size, step)
		for s := range wins {
			fv := BedEntry[string]{
				ChrSpan: ChrSpan{fe.FaHeader(), Span{s.Start, s.End}},
				Fields:  fe.FaSeq()[s.Start:s.End],
			}
			if !yield(fv) {
				return
			}
		}
	}
}

func FaWins[F FaEnter](fa iter.Seq2[F, error], size int64, step int64) iter.Seq2[BedEntry[string], error] {
	return func(yield func(BedEntry[string], error) bool) {
		for fe, err := range fa {
			if err != nil {
				if !yield(BedEntry[string]{}, err) {
					return
				}
			}
			wins := FaEntryWins(fe, size, step)
			for view := range wins {
				if !yield(view, nil) {
					return
				}
			}
		}
	}
}
