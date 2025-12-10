package fastats

import (
	"fmt"
	"iter"
)

func CollectChrSpannerMap[C ChrSpanner](cit iter.Seq2[C, error]) (map[string][]C, error) {
	m := map[string][]C{}
	for c, err := range cit {
		if err != nil {
			return nil, err
		}
		m[c.SpanChr()] = append(m[c.SpanChr()], c)
	}
	return m, nil
}

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

func NameSpan(chr string, start, end int64) string {
	return fmt.Sprintf("%v:%v-%v", chr, start, end)
}

func ExtractOne(f FaEntry, s Span) (FaEntry, error) {
	if s.Start < 0 || s.Start >= int64(len(f.Seq)) {
		return FaEntry{}, fmt.Errorf("ExtractOne: s.Start %v out of range of len(f.Seq) %v", s.Start, len(f.Seq))
	}
	if s.End < 0 || s.End > int64(len(f.Seq)) {
		return FaEntry{}, fmt.Errorf("ExtractOne: s.End %v out of range of len(f.Seq) %v", s.End, len(f.Seq))
	}
	return FaEntry{
		Header: NameSpan(f.Header, s.Start, s.End),
		Seq:    f.Seq[s.Start:s.End],
	}, nil
}

type ChrSpanFa[C ChrSpanner, F FaEnter] struct {
	ChrSpanner C
	FaEnter F
}

func ExtractChrSpanFa[F FaEnter, C ChrSpanner](fit iter.Seq2[F, error], cit iter.Seq2[C, error]) iter.Seq2[ChrSpanFa[C, F], error] {
	return func(yield func(ChrSpanFa[C, F], error) bool) {
		m, e := CollectChrSpannerMap(cit)
		if e != nil {
			if !yield(ChrSpanFa[C,F]{}, e) {
				return
			}
		}
		for f, err := range fit {
			if err != nil {
				if !yield(ChrSpanFa[C,F]{}, e) {
					return
				}
			}
			spans := m[f.FaHeader()]
			for _, span := range spans {
				out := ChrSpanFa[C, F]{ChrSpanner: span, FaEnter: f}
				if !yield(out, nil) {
					return
				}
			}
		}
	}
}

func ExtractFasta[F FaEnter, C ChrSpanner](fit iter.Seq2[F, error], cit iter.Seq2[C, error]) iter.Seq2[FaEntry, error] {
	return func(yield func(FaEntry, error) bool) {
		for csf, e := range ExtractChrSpanFa(fit, cit) {
			if e != nil {
				if !yield(FaEntry{}, e) {
					return
				}
				continue
			}
			fa, e := ExtractOne(ToFaEntry(csf.FaEnter), ToSpan(csf.ChrSpanner))
			if !yield(fa, e) {
				return
			}
		}
	}
}
