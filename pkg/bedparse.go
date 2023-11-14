package fastats

import (
	"fmt"
	"io"
	"encoding/csv"
)

type Span struct {
	Start int64
	End int64
}

type ChrSpan struct {
	Chr string
	Span
}

type BedEntry[FieldsT any] struct {
	ChrSpan
	Fields FieldsT
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

func ParseBed[FT any](r io.Reader, fieldParse func([]string) (FT, error)) *Iterator[BedEntry[FT]] {
	return &Iterator[BedEntry[FT]]{Iteratef: func(yield func(BedEntry[FT]) error) error {
		cr := csv.NewReader(r)
		cr.LazyQuotes = true
		cr.Comma = rune('\t')
		cr.ReuseRecord = true
		cr.FieldsPerRecord = -1

		for l, e := cr.Read(); e != io.EOF; l, e = cr.Read() {
			b, e := ParseBedEntry(l, fieldParse)
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

func ParseBedFlat(r io.Reader) *Iterator[BedEntry[[]string]] {
	return ParseBed(r, func(fields []string) ([]string, error) {
		out := make([]string, len(fields))
		copy(out, fields)
		return out, nil
	})
}
