package fastats

import (
	"fmt"
	"io"
	"encoding/csv"
	"iter"
)

type Span struct {
	Start int64
	End int64
}

func (s Span) SpanStart() int64 {
	return s.Start
}

func (s Span) SpanEnd() int64 {
	return s.End
}

type Spanner interface {
	SpanStart() int64
	SpanEnd() int64
}

func StartOneIndex[S Spanner](s S) int64 {
	return s.SpanStart() + 1
}

func ToSpan[S Spanner](s S) Span {
	return Span{Start: s.SpanStart(), End: s.SpanEnd()}
}

type ChrSpan struct {
	Chr string
	Span
}

func ToChrSpan[C ChrSpanner](c C) ChrSpan {
	return ChrSpan{Chr: c.SpanChr(), Span: ToSpan(c)}
}

func (c ChrSpan) SpanChr() string {
	return c.Chr
}

type ChrSpanner interface {
	Spanner
	SpanChr() string
}

type BedEntry[FieldsT any] struct {
	ChrSpan
	Fields FieldsT
}

func (b BedEntry[T]) BedFields() T {
	return b.Fields
}

type BedEnter[FieldsT any] interface {
	ChrSpanner
	BedFields() FieldsT
}

func ToBedEntry[B BedEnter[FieldsT], FieldsT any](b B) BedEntry[FieldsT] {
	if val, ok := any(b).(BedEntry[FieldsT]); ok {
		return val
	}
	return BedEntry[FieldsT] {
		ChrSpan: ToChrSpan(b),
		Fields: b.BedFields(),
	}
}

func ScanOne(field string, ptr any) (n int, err error) {
	if s, ok := ptr.(*string); ok {
		*s = field
		return 1, nil
	}
	nread, e := fmt.Sscanf(field, "%v", ptr)
	return nread, e
}

func ScanOneDot(field string, ptr any) (n int, err error) {
	if ptr == nil {
		return 1, nil
	}
	n, e := ScanOne(field, ptr)
	if e == nil {
		return n, nil
	}
	if field == "." {
		return 1, nil
	}
	return n, fmt.Errorf("ScanOneDot: field: %v; ptr: %v; err: %w", field, ptr, e)
}

func ScanF(f func(string, any) (int, error), line []string, ptrs ...any) (n int, err error) {
	for i, ptr := range ptrs {
		nread, e := f(line[i], ptr)
		n += nread
		if e != nil {
			return n, e
		}
	}
	return n, nil
}

func Scan(line []string, ptrs ...any) (n int, err error) {
	return ScanF(ScanOne, line, ptrs...)
}

func ScanDot(line []string, ptrs ...any) (n int, err error) {
	return ScanF(ScanOneDot, line, ptrs...)
}

func ParseBedEntry[FT any](line []string, fieldParse func([]string) (FT, error)) (BedEntry[FT], error) {
	var b BedEntry[FT]
	if len(line) < 3 {
		return b, fmt.Errorf("ParseBedEntry: len(line) %v < 3", len(line))
	}

	_, e := Scan(line[:3], &b.Chr, &b.Start, &b.End)
	if e != nil {
		return b, e
	}

	b.Fields, e = fieldParse(line[3:])
	return b, e
}

func ParseBed[FT any](r io.Reader, fieldParse func([]string) (FT, error)) iter.Seq2[BedEntry[FT], error] {
	return func(yield func(BedEntry[FT], error) bool) {
		cr := csv.NewReader(r)
		cr.LazyQuotes = true
		cr.Comma = rune('\t')
		cr.ReuseRecord = true
		cr.FieldsPerRecord = -1

		for l, e := cr.Read(); e != io.EOF; l, e = cr.Read() {
			b, e := ParseBedEntry(l, fieldParse)
			ok := yield(b, e)
			if e != nil || !ok {
				return
			}
		}
	}
}

func ParseBedFlat(r io.Reader) iter.Seq2[BedEntry[[]string], error] {
	return ParseBed(r, func(fields []string) ([]string, error) {
		out := make([]string, len(fields))
		copy(out, fields)
		return out, nil
	})
}

func SpreadBed[B BedEnter[T], T any](it iter.Seq2[B, error]) func(func(BedEntry[T], error) bool) {
	return func(yield func(BedEntry[T], error) bool) {
		it(func(b B, e error) bool {
			for i := b.SpanStart(); i < b.SpanEnd(); i++ {
				sub := BedEntry[T]{}
				sub.Chr = b.SpanChr()
				sub.Start = i
				sub.End = i+1
				sub.Fields = b.BedFields()
				ok := yield(sub, e)
				return ok && e == nil
			}
			return true
		})
	}
}
