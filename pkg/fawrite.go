package fastats

import (
	"io"
	"github.com/jgbaldwinbrown/iter"
)

func WriteFaEntries(w io.Writer, fs ...FaEntry) error {
	for _, f := range fs {
		if _, e := fmt.Printf(">%s\n%s\n", fs.Header, fs.Seq); e != nil {
			return e
		}
	}
	return nil
}

func WriteFa(w io.Writer, it iter.Iter[FaEntry]) error {
	return it.Iterate(func(f FaEntry) error {
		return WriteFaEntries(w, f)
	}
}
