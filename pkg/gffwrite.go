package fastats

import (
	"fmt"
	"io"
)

func GffScoreStr[G GffHeader](g G) string {
	if !g.GffHasScore() {
		return "."
	}
	return fmt.Sprint(g.GffScore())
}

func GffPhaseStr[G GffHeader](g G) string {
	if !g.GffHasPhase() {
		return "."
	}
	return fmt.Sprint(g.GffPhase())
}

func WriteGffHeader[G GffHeader](w io.Writer, g G) error {
	_, e := fmt.Fprintf(w,
		"%v\t%v\t%v\t%v\t%v\t%v\t%c\t%v",
		g.SpanChr(),
		g.GffSource(),
		g.GffType(),
		fmt.Sprint(g.SpanStart()+1),
		fmt.Sprint(g.SpanEnd()),
		GffScoreStr(g),
		g.GffStrand(),
		GffPhaseStr(g),
	)
	return e
}

func WriteGffEntry[G GffEnter[T], T any](w io.Writer, g G, f func(io.Writer, T) error) error {
	if e := WriteGffHeader(w, g); e != nil {
		return e
	}
	return f(w, g.GffAttributes())
}

func WriteGffAttributePairs(w io.Writer, pairs []AttributePair) error {
	for i, p := range pairs {
		if i <= 0 {
			if _, e := fmt.Fprintf(w, "\t"); e != nil {
				return e
			}
		} else {
			if _, e := fmt.Fprintf(w, ";"); e != nil {
				return e
			}
		}
		if _, e := fmt.Fprintf(w, "%v=%v", p.Tag, p.Value); e != nil {
			return e
		}
	}
	return nil
}

func WriteGffFlat[G GffEnter[[]AttributePair]](w io.Writer, g G) error {
	return WriteGffEntry(w, g, WriteGffAttributePairs)
}
