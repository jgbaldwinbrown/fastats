package fastats

import (
	"strings"
	"fmt"
	"encoding/csv"
	"io"
	"iter"
)

func ParseVcfMainFields[T any](v *VcfEntry[T], line []string) error {
	if len(line) < 7 {
		return fmt.Errorf("ParseVcfMainFields: len(line) %v < 7", len(line))
	}

	_, e := ScanDot(line[:7], &v.Chr, &v.Start, &v.ID, &v.Ref, nil, &v.Qual, &v.Filter)
	if e != nil {
		return e
	}
	v.Start--
	v.End = v.Start + 1

	v.Alts = strings.Split(line[4], ",")

	return nil
}

func ParseVcfEntry[T any](line []string, f func(line []string) (T, error)) (VcfEntry[T], error) {
	var v VcfEntry[T]

	e := ParseVcfMainFields(&v, line)
	if e != nil {
		return v, e
	}

	v.InfoAndSamples, e = f(line)
	if e != nil {
		return v, e
	}

	return v, nil
}

func ParseVcf[T any](r io.Reader, f func(line []string) (T, error)) iter.Seq2[VcfEntry[T], error] {
	return func(yield func(VcfEntry[T], error) bool) {
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
			if !yield(b, e) {
				return
			}
		}
	}
}

