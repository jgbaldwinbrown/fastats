package fastats

import (
	"os"
	"bufio"
	"fmt"
	"github.com/jgbaldwinbrown/iter"
)

func ToFa(it iter.Iter[FqEntry]) *iter.Iterator[FaEntry] {
	return iter.Transform(it, func(f FqEntry) (FaEntry, error) {
		return f.FaEntry, nil
	})
}

func FullToFa() {
	it := ToFa(ParseFastq(os.Stdin))
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	err := it.Iterate(func(f FaEntry) error {
		_, e := fmt.Fprintf(w, ">%v\n%v\n", f.Header, f.Seq)
		return e
	})
	if err != nil {
		panic(err)
	}
}
