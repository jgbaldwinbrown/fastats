package fastats

import (
	"iter"
)

func BedWinWithChrs[C, B ChrSpanner](chroms iter.Seq[C], bed iter.Seq[B], winsize, winstep int64) map[ChrSpan][]B {
	m := map[ChrSpan][]B{}
	for chrSpan := range chroms {
		chr := ToChrSpan(chrSpan)
		for i := (chr.Start / winstep) * winstep; i < chr.End; i += winstep {
			m[ChrSpan{Chr: chr.Chr, Span: Span{Start: i, End: i + winsize}}] = []B{}
		}
	}
	for b := range bed {
		cs := ToChrSpan(b)
		for i := (cs.Start / winstep) * winstep; i < cs.End; i += winstep {
			winCs := ChrSpan{Chr: cs.Chr, Span: Span{Start: i, End: i + winsize}}
			if winVals, ok := m[winCs]; ok {
				winVals = append(winVals, b)
				m[winCs] = winVals
			}
		}
	}
	return m
}

func BedWinImplicitChrs[B ChrSpanner](bed iter.Seq[B], winsize, winstep int64) map[ChrSpan][]B {
	chrs := map[string]Span{}
	for b := range bed {
		bc := ToChrSpan(b)
		s, ok := chrs[bc.Chr]
		if !ok || bc.Start < s.Start {
			s.Start = bc.Start
		}
		if !ok || bc.End > s.End {
			s.End = bc.End
		}
		chrs[bc.Chr] = s
	}
	chrsbed := func(y func(ChrSpan) bool) {
		for chr, span := range chrs {
			if !y(ChrSpan{Chr: chr, Span: span}) {
				return
			}
		}
	}
	return BedWinWithChrs(chrsbed, bed, winsize, winstep)
}
