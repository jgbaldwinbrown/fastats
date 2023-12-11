package fastats

import (
	"fmt"
	"encoding/csv"
	"io"
	"github.com/jgbaldwinbrown/iter"
)

func ParseVcfEntry[T any](line []string, f func(line []string) (T, error)) (VcfEntry[T], error) {
	var v VcfEntry[T]
	v.Alts = make([]string, 1)
	if len(line) < 7 {
		return v, fmt.Errorf("ParseSimpleVcfEntry: len(line) %v < 7", len(line))
	}
	_, e := ScanDot(line[:7], &v.Chr, &v.Start, &v.ID, &v.Ref, &v.Alts[0], &v.Qual, &v.Filter)
	if e != nil {
		return v, e
	}
	v.Start--
	v.End = v.Start + 1

	v.InfoAndSamples, e = f(line)
	if e != nil {
		return v, e
	}

	return v, nil
}

func ParseVcf[T any](r io.Reader, f func(line []string) (T, error)) *iter.Iterator[VcfEntry[T]] {
	return &iter.Iterator[VcfEntry[T]]{Iteratef: func(yield func(VcfEntry[T]) error) error {
		cr := csv.NewReader(r)
		cr.LazyQuotes = true
		cr.Comma = rune('\t')
		cr.ReuseRecord = true
		cr.FieldsPerRecord = -1

		for l, e := cr.Read(); e != io.EOF; l, e = cr.Read() {
			if len(l) < 7 {
				continue
			}
			if len(l) > 0 && commentRe.MatchString(l[0]) {
				continue
			}

			b, e := ParseVcfEntry(l, f)
			if e != nil {
				return e
			}
			e = yield(b)
			if e != nil {
				return e
			}
		}

		return nil
	}}
}

