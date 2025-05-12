package fastats

import (
	"encoding/csv"
	"fmt"
	"io"
	"iter"
	"strings"
	"strconv"
)

func ParseVcfMainFields[T any](v *VcfEntry[T], line []string) error {
	if len(line) < 7 {
		return fmt.Errorf("ParseVcfMainFields: len(line) %v < 7", len(line))
	}

	var e error
	v.Chr = line[0]
	if v.Start, e = strconv.ParseInt(line[1], 0, 64); e != nil {
		return fmt.Errorf("ParseVcfMainFields: %w", e)
	}
	v.Start--
	v.End = v.Start + 1

	v.ID = line[2]
	v.Ref = line[3]
	v.Alts = strings.Split(line[4], ",")
	if v.Qual, e = strconv.Atoi(line[5]); e != nil {
		if line[5] != "." {
			return fmt.Errorf("ParseVcfMainFields: %w", e)
		}
		v.Qual = 0
	}
	v.Filter = line[6]

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

func ParseVcfFlat(r io.Reader) iter.Seq2[VcfEntry[[]string], error] {
	return ParseVcf(r, func(line []string) ([]string, error) {
		if len(line) <= 7 {
			return []string{}, nil
		}
		return line[7:], nil
	})
}
