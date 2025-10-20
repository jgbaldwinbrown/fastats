package fastats

import (
	"io"
	"fmt"
)

func WriteBedHeader[B ChrSpanner](w io.Writer, b B) error {
	_, e := fmt.Fprintf(w, "%v\t%v\t%v", b.SpanChr(), b.SpanStart(), b.SpanEnd())
	return e
}

func WriteBedEntry[B BedEnter[T], T any](w io.Writer, b B, writeFields func(io.Writer, T) error) error {
	if e := WriteBedHeader(w, b); e != nil {
		return e
	}
	if e := writeFields(w, b.BedFields()); e != nil {
		return e
	}
	if _, e := fmt.Fprintf(w, "\n"); e != nil {
		return e
	}
	return nil
}

func writeTabbedSlice[S ~[]T, T any](w io.Writer, s S) error {
	for _, field := range s {
		if _, e := fmt.Fprintf(w, "\t%v", field); e != nil {
			return e
		}
	}
	return nil
}

func WriteBedEnterFlat[B BedEnter[[]T], T any](w io.Writer, b B) error {
	return WriteBedEntry(w, b, writeTabbedSlice[[]T, T])
}

func WriteBedEnterSingle[B BedEnter[T], T any](w io.Writer, b B) error {
	return WriteBedEntry(w, b, func(w2 io.Writer, t T) error {
		_, e := fmt.Fprintf(w, "\t%v", t)
		return e
	})
}
