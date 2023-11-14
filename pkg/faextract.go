package fastats

import (
	"fmt"
	"github.com/jgbaldwinbrown/iter"
)

func CollectChrSpanMap(cit iter.Iter[ChrSpan]) (map[string][]Span, error) {
	m := map[string][]Span{}
	err := cit.Iterate(func(c ChrSpan) error {
		m[c.Chr] = append(m[c.Chr], c.Span)
		return nil
	})
	if err != nil {
		return nil, err
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

func ExtractFasta(fit iter.Iter[FaEntry], cit iter.Iter[ChrSpan]) *iter.Iterator[FaEntry] {
	return &iter.Iterator[FaEntry]{Iteratef: func(yield func(FaEntry) error) error {
		m, e := CollectChrSpanMap(cit)
		if e != nil {
			return e
		}
		fit.Iterate(func(f FaEntry) error {
			spans := m[f.Header]
			for _, span := range spans {
				out, e := ExtractOne(f, span)
				if e != nil {
					return e
				}
				e = yield(out)
				if e != nil {
					return e
				}
			}
			return nil
		})
		return nil
	}}
}
