package fastats

import (
	"fmt"
	"iter"
)

func CollectChrSpanMap[C ChrSpanner](cit iter.Seq2[C, error]) (map[string][]Span, error) {
	m := map[string][]Span{}
	for c, err := range cit {
		if err != nil {
			return nil, err
		}
		m[c.SpanChr()] = append(m[c.SpanChr()], Span{c.SpanStart(), c.SpanEnd()})
	}
	return m, nil
}

func ExtractOne(f FaEntry, s Span) (FaEntry, error) {
	if s.Start < 0 || s.Start >= int64(len(f.Seq)) {
		return FaEntry{}, fmt.Errorf("ExtractOne: s.Start %v out of range of len(f.Seq) %v", s.Start, len(f.Seq))
	}
	if s.End < 0 || s.End > int64(len(f.Seq)) {
		return FaEntry{}, fmt.Errorf("ExtractOne: s.End %v out of range of len(f.Seq) %v", s.End, len(f.Seq))
	}
	return FaEntry{
		Header: fmt.Sprintf("%v:%v-%v", f.Header, s.Start, s.End),
		Seq: f.Seq[s.Start:s.End],
	}, nil
}

func ExtractFasta[F FaEnter, C ChrSpanner](fit iter.Seq2[F, error], cit iter.Seq2[C, error]) iter.Seq2[FaEntry, error] {
	return func(yield func(FaEntry, error) bool) {
		m, e := CollectChrSpanMap(cit)
		if e != nil {
			if !yield(FaEntry{}, e) {
				return
			}
		}
		for f, err := range fit {
			if err != nil {
				if !yield(FaEntry{}, e) {
					return
				}
			}
			spans := m[f.FaHeader()]
			for _, span := range spans {
				out, e := ExtractOne(toFaEntry(f), span)
				if !yield(out, e) {
					return
				}
			}
		}
	}
}
