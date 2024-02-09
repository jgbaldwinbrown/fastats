package fastats

import (
	"fmt"
	"os"
	"bufio"
	"strings"
	"io"
	"github.com/jgbaldwinbrown/iter"
)

type FaEntry struct {
	Header string
	Seq string
}

func parseFasta(r io.Reader, yield func(f FaEntry) error) error {
	s := bufio.NewScanner(r)
	s.Buffer([]byte{}, 1e12)

	started := false
	var hdr string
	var seq strings.Builder

	for s.Scan() {
		if s.Err() != nil {
			return s.Err()
		}
		if len(s.Text()) < 1 {
			continue
		}

		if s.Text()[0] == '>' {
			// fmt.Println("found header:", s.Text())
			if started {
				e := yield(FaEntry{Header: hdr, Seq: seq.String()})
				if e != nil {
					return e
				}
			}
			hdr = s.Text()[1:]
			seq.Reset()
			started = true
			continue
		}

		_, e := seq.WriteString(s.Text())
		if e != nil {
			return e
		}
	}

	if started {
		e := yield(FaEntry{Header: hdr, Seq: seq.String()})
		if e != nil {
			return e
		}
	}

	return nil
}

func ParseFasta(r io.Reader) *iter.Iterator[FaEntry] {
	it := func(yield func(FaEntry) error) error {
		return parseFasta(r, yield)
	}
	return &iter.Iterator[FaEntry]{Iteratef: it}
}

type FaLen struct {
	Name string
	Len int64
}

func Chrlen(f FaEntry) FaLen {
		return FaLen{f.Header, int64(len(f.Seq))}
}

func Chrlens(it iter.Iter[FaEntry]) *iter.Iterator[FaLen] {
	itf := func(yield func(FaLen) error) error {
		return it.Iterate(func(f FaEntry) error {
			return yield(Chrlen(f))
		})
	}
	return &iter.Iterator[FaLen]{Iteratef: itf}
}

func PrintFaLen(w io.Writer, l FaLen) error {
	_, e := fmt.Printf("%v\t%v\n", l.Name, l.Len)
	return e
}

func FullChrlens() {
	fa := ParseFasta(os.Stdin)
	lens := Chrlens(fa)
	err := lens.Iterate(func(l FaLen) error {
		return PrintFaLen(os.Stdout, l)
	})
	if err != nil {
		panic(err)
	}
}
