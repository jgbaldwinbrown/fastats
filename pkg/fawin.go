package fastats

import (
	"github.com/jgbaldwinbrown/iter"
)

func Wins(start, end, size, step int64) *iter.Iterator[Span] {
	return &iter.Iterator[Span]{Iteratef: func(yield func(Span) error) error {
		var i int64
		for i = start; i < end; i += step {
			e := yield(Span{i, i + step})
			if e != nil {
				return e
			}
		}
		return nil
	}}
}

type FaView struct {
	FaEntry
	Span
}

func FaEntryWins(fe FaEntry, size int64, step int64) *iter.Iterator[FaView] {
	return &iter.Iterator[FaView]{Iteratef: func(yield func(FaView) error) error {
		wins := Wins(0, int64(len(fe.Seq)), size, step)
		return wins.Iterate(func(s Span) error {
			fv := FaView {fe, s}
			return yield(fv)
		})
	}}
}
