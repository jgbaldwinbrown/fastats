package fastats

import (
	"github.com/jgbaldwinbrown/iter"
)

func Wins(start, end, size, step int64) *iter.Iterator[Span] {
	return &iter.Iterator[Span]{Iteratef: func(yield func(Span) error) error {
		var i int64
		for i = start; i < end; i += step {
			winend := i + size
			if winend > end {
				winend = end
			}
			e := yield(Span{i, winend})
			if e != nil {
				return e
			}
		}
		return nil
	}}
}

func FaEntryWins(fe FaEntry, size int64, step int64) *iter.Iterator[BedEntry[string]] {
	return &iter.Iterator[BedEntry[string]]{Iteratef: func(yield func(BedEntry[string]) error) error {
		wins := Wins(0, int64(len(fe.Seq)), size, step)
		return wins.Iterate(func(s Span) error {
			fv := BedEntry[string]{
				ChrSpan: ChrSpan{fe.Header, Span{s.Start, s.End}},
				Fields: fe.Seq[s.Start : s.End],
			}
			return yield(fv)
		})
	}}
}

func FaWins(fa iter.Iter[FaEntry], size int64, step int64) *iter.Iterator[BedEntry[string]] {
	return &iter.Iterator[BedEntry[string]]{Iteratef: func(yield func(BedEntry[string]) error) error {
		return fa.Iterate(func(fe FaEntry) error {
			wins := FaEntryWins(fe, size, step)
			return wins.Iterate(func(view BedEntry[string]) error {
				return yield(view)
			})
		})
	}}
}
