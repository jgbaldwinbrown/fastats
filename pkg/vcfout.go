package fastats

import (
	"encoding/csv"
	"fmt"
	"io"
	"iter"
	"regexp"
	"strings"
)

type VcfHead struct {
	ChrSpan
	ID     string
	Ref    string
	Alts   []string
	Qual   int
	Filter string
}

func (v VcfHead) VcfID() string     { return v.ID }
func (v VcfHead) VcfRef() string    { return v.Ref }
func (v VcfHead) VcfAlts() []string { return v.Alts }
func (v VcfHead) VcfQual() int      { return v.Qual }
func (v VcfHead) VcfFilter() string { return v.Filter }

type VcfHeader interface {
	ChrSpanner
	VcfID() string
	VcfRef() string
	VcfAlts() []string
	VcfQual() int
	VcfFilter() string
}

func ToVcfHead[V VcfHeader](v V) VcfHead {
	if ptr, ok := any(&v).(*VcfHead); ok {
		return *ptr
	}
	return VcfHead{
		ChrSpan: ToChrSpan(v),
		ID:      v.VcfID(),
		Ref:     v.VcfRef(),
		Alts:    v.VcfAlts(),
		Qual:    v.VcfQual(),
		Filter:  v.VcfFilter(),
	}
}

type VcfEntry[T any] struct {
	VcfHead
	InfoAndSamples T
}

func (v VcfEntry[T]) VcfInfoAndSamples() T {
	return v.InfoAndSamples
}

type VcfEnter[T any] interface {
	VcfHeader
	VcfInfoAndSamples() T
}

func ToVcfEntry[V VcfEnter[T], T any](v V) VcfEntry[T] {
	if ptr, ok := any(&v).(*VcfEntry[T]); ok {
		return *ptr
	}
	return VcfEntry[T]{
		VcfHead:        ToVcfHead(v),
		InfoAndSamples: v.VcfInfoAndSamples(),
	}
}

type InfoPair[T any] struct {
	Key string
	Val T
}

type Formatter interface {
	Format(format string) string
}

type SampleSet[T Formatter] struct {
	Format  []string
	Samples [][]T
}

type StructuredInfoSamples[InfoT any, SampleT Formatter] struct {
	Info    []InfoPair[InfoT]
	Samples SampleSet[SampleT]
}

func InfoToString[T any](is []InfoPair[T]) (string, error) {
	var b strings.Builder

	if len(is) < 0 {
		return "", nil
	}

	if len(is) > 0 {
		_, e := fmt.Fprintf(&b, "%v=%v", is[0].Key, is[0].Val)
		if e != nil {
			return "", e
		}
	}

	if len(is) < 2 {
		return b.String(), nil
	}

	for _, info := range is[1:] {
		_, e := fmt.Fprintf(&b, ";%v=%v", info.Key, info.Val)
		if e != nil {
			return "", e
		}
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
		if e != nil {
			return "", e
		}
	}
	for i := 1; i < len(sample); i++ {
		samp := sample[i]
		form := format[i]
		_, e := fmt.Fprintf(&b, ":%v", samp.Format(form))
		if e != nil {
			return "", e
		}
	}

	return b.String(), nil
}

func AppendSamples[T Formatter](out []string, s SampleSet[T]) ([]string, error) {
	out = append(out, strings.Join(s.Format, ":"))
	for _, samp := range s.Samples {
		str, err := FormatSample[T](s.Format, samp)
		if err != nil {
			return nil, err
		}

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

func ParseVcfHead(line []string) (VcfHead, error) {
	var v VcfEntry[struct{}]
	e := ParseVcfMainFields(&v, line)
	return v.VcfHead, e
}

func ParseSimpleVcfEntry(line []string) (VcfEntry[struct{}], error) {
	v, e := ParseVcfHead(line)
	return VcfEntry[struct{}]{VcfHead: v}, e
}

var commentRe = regexp.MustCompile(`^#`)

func ParseSimpleVcf(r io.Reader) iter.Seq2[VcfEntry[struct{}], error] {
	return func(yield func(VcfEntry[struct{}], error) bool) {
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
			if !yield(b, e) {
				return
			}
		}
	}
}
