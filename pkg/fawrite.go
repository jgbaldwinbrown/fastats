package fastats

import (
	"fmt"
	"io"
	"iter"
)

func WriteFaEntries[F FaEnter](w io.Writer, fs ...F) error {
	for _, f := range fs {
		if _, e := fmt.Fprintf(w, ">%s\n%s\n", f.FaHeader(), f.FaSeq()); e != nil {
			return e
		}
	}
	return nil
}

func WriteFa[F FaEnter](w io.Writer, it iter.Seq2[F, error]) error {
	for f, e := range it {
		if e != nil {
			return e
		}
		if e := WriteFaEntries(w, f); e != nil {
			return e
		}
	}
	return nil
}
