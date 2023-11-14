package fastats

import (
	"regexp"
	"encoding/csv"
	"io"
	"fmt"
	"strings"
)

type VcfEntry[T any] struct {
	ChrSpan
	ID string
	Ref string
	Alts []string
	Qual int
	Filter string
	InfoAndSamples T
}

type InfoPair[T any] struct {
	Key string
	Val T
}

type Formatter interface {
	Format(format string) string
}

type SampleSet[T Formatter] struct {
	Format []string
	Samples [][]T
}

type StructuredInfoSamples[InfoT any, SampleT Formatter] struct {
	Info []InfoPair[InfoT]
	Samples SampleSet[SampleT]
}

func InfoToString[T any](is []InfoPair[T]) (string, error) {
	var b strings.Builder

	if len(is) < 0 {
		return "", nil
	}

	if len(is) > 0 {
		_, e := fmt.Fprintf(&b, "%v=%v", is[0].Key, is[0].Val)
		if e != nil { return "", e }
	}

	if len(is) < 2 {
		return b.String(), nil
	}

	for _, info := range is[1:] {
		_, e := fmt.Fprintf(&b, ";%v=%v", info.Key, info.Val)
		if e != nil { return "", e }
	}

	return b.String(), nil
}

func FormatSample[T Formatter, S ~[]T](format []string, sample S) (string, error) {
	if len(format) < len(sample) {
		return "", fmt.Errorf("len(format) %v < len(sample) %v", len(format), len(sample))
	}

	var b strings.Builder

	if len(sample) > 0 {
		_, e := fmt.Fprintf(&b, "%v", sample[0].Format(format[0]))
		if e != nil { return "", e }
	}
	for i := 1; i < len(sample); i++ {
		samp := sample[i]
		form := format[i]
		_, e := fmt.Fprintf(&b, ":", samp.Format(form))
		if e != nil { return "", e }
	}

	return b.String(), nil
}

func AppendSamples[T Formatter](out []string, s SampleSet[T]) ([]string, error) {
	out = append(out, strings.Join(s.Format, ":"))
	for _, samp := range s.Samples {
		str, err := FormatSample[T](s.Format, samp)
		if err != nil { return nil, err }

		out = append(out, str)
	}
	return out, nil
}

func StructuredVcfEntryToCsv[InfoT any, SampleT Formatter](buf []string, v VcfEntry[StructuredInfoSamples[InfoT, SampleT]]) ([]string, error) {
	buf = append(buf[:0], v.Chr, fmt.Sprintf("%v", v.Start+1), v.ID, v.Ref, strings.Join(v.Alts, ","), fmt.Sprintf("%v", v.Qual), v.Filter)

	info, err := InfoToString(v.InfoAndSamples.Info)
	if err != nil {
		return nil, err
	}
	buf = append(buf, info)

	buf, err = AppendSamples(buf, v.InfoAndSamples.Samples)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func ParseSimpleVcfEntry(line []string) (VcfEntry[struct{}], error) {
	var v VcfEntry[struct{}]
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
	return v, nil
}

var commentRe = regexp.MustCompile(`^#`)

func ParseSimpleVcf(r io.Reader) *Iterator[VcfEntry[struct{}]] {
	return &Iterator[VcfEntry[struct{}]]{Iteratef: func(yield func(VcfEntry[struct{}]) error) error {
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

			b, e := ParseSimpleVcfEntry(l)
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

