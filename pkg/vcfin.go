package fastats

import (
	"regexp"
	"errors"
	"encoding/csv"
	"fmt"
	"io"
	"iter"
	"slices"
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
	if v.Qual, e = strconv.ParseFloat(line[5], 64); e != nil {
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

func ParseVcfCore[T any](cr *csv.Reader, f func(line []string) (T, error)) iter.Seq2[VcfEntry[T], error] {
	return func(yield func(VcfEntry[T], error) bool) {
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

func ParseVcf[T any](r io.Reader, f func(line []string) (T, error)) iter.Seq2[VcfEntry[T], error] {
	cr := csv.NewReader(r)
	cr.LazyQuotes = true
	cr.Comma = rune('\t')
	cr.ReuseRecord = true
	cr.FieldsPerRecord = -1

	return ParseVcfCore(cr, f)
}

var chromRe = regexp.MustCompile(`^#CHROM`)

func ParseVcfPlusHeader[T any](r io.Reader, f func(line []string) (T, error)) ([]string, iter.Seq2[VcfEntry[T], error], error) {
	cr := csv.NewReader(r)
	cr.LazyQuotes = true
	cr.Comma = rune('\t')
	cr.ReuseRecord = true
	cr.FieldsPerRecord = -1
	var header []string

	for l, e := cr.Read(); e != io.EOF; l, e = cr.Read() {
		if len(l) > 0 && chromRe.MatchString(l[0]) {
			header = slices.Clone(l)
			break
		}
		if len(l) < 7 {
			continue
		}
		if len(l) > 0 && commentRe.MatchString(l[0]) {
			continue
		}
	}
	return header, ParseVcfCore(cr, f), nil
}

func ParseVcfFlat(r io.Reader) iter.Seq2[VcfEntry[[]string], error] {
	return ParseVcf(r, func(line []string) ([]string, error) {
		if len(line) <= 7 {
			return []string{}, nil
		}
		return slices.Clone(line[7:]), nil
	})
}

type StandardVcfInfoAndSamples struct {
	InfoKeys []string
	InfoVals []string
	Format []string
	Samples [][]string
}

var ErrVcfFormat = errors.New("Vcf format error")

func ParseInfo(info string) (keys, vals []string) {
	fields := strings.Split(info, ";")
	keys = make([]string, 0, len(fields))
	vals = make([]string, 0, len(fields))
	for _, field := range fields {
		key, val, _ := strings.Cut(field, "=")
		keys = append(keys, key)
		vals = append(vals, val)
	}
	return keys, vals
}

func ParseStandardVcfInfoAndSamples(line []string) (StandardVcfInfoAndSamples, error) {
	if len(line) < 7 {
		return StandardVcfInfoAndSamples{}, nil
	}
	var s StandardVcfInfoAndSamples
	s.InfoKeys, s.InfoVals = ParseInfo(line[7])
	s.Format = strings.Split(line[8], ":")
	for i := 9; i < len(line); i++ {
		s.Samples = append(s.Samples, strings.Split(line[i], ","))
	}
	for i := 9; i < len(line); i++ {
		if len(s.Samples[len(s.Samples)-1]) != len(s.Format) {
			return s, fmt.Errorf("%w: len(s.Samples[%v]) %v, %v != len(s.Format) %v, %v", ErrVcfFormat, i, len(s.Samples[i]), s.Samples[i], len(s.Format), s.Format)
		}
	}
	return s, nil
}
