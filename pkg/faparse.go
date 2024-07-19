package fastats

import (
	"log"
	"fmt"
	"os"
	"bufio"
	"strings"
	"io"
	"iter"
)

type FaEntry struct {
	Header string
	Seq string
}

func (f FaEntry) FaHeader() string { return f.Header }
func (f FaEntry) FaSeq() string { return f.Seq }

type FaEnter interface {
	FaHeader() string
	FaSeq() string
}

func parseFasta(r io.Reader, yield func(FaEntry, error) bool) {
	s := bufio.NewScanner(r)
	s.Buffer([]byte{}, 1e12)

	started := false
	var hdr string
	var seq strings.Builder

	for s.Scan() {
		if s.Err() != nil {
			if !yield(FaEntry{}, s.Err()) {
				return
			}
		}
		if len(s.Text()) < 1 {
			continue
		}

		if s.Text()[0] == '>' {
			// fmt.Println("found header:", s.Text())
			if started {
				if !yield(FaEntry{Header: hdr, Seq: seq.String()}, nil) {
					return
				}
			}
			hdr = s.Text()[1:]
			seq.Reset()
			started = true
			continue
		}

		_, e := seq.WriteString(s.Text())
		if e != nil {
			if !yield(FaEntry{}, e) {
				return
			}
		}
	}

	if started {
		yield(FaEntry{Header: hdr, Seq: seq.String()}, nil)
	}
}

func ParseFasta(r io.Reader) iter.Seq2[FaEntry, error] {
	return func(yield func(FaEntry, error) bool) {
		parseFasta(r, yield)
	}
}

type FaLen struct {
	Name string
	Len int64
}

func Chrlen[F FaEnter](f F) FaLen {
		return FaLen{f.FaHeader(), int64(len(f.FaSeq()))}
}

func Chrlens[F FaEnter](it iter.Seq2[F, error]) iter.Seq2[FaLen, error] {
	return func(yield func(FaLen, error) bool) {
		for f, e := range it {
			if !yield(Chrlen(f), e) {
				return
			}
		}
	}
}

func PrintFaLen(w io.Writer, l FaLen) error {
	_, e := fmt.Printf("%v\t%v\n", l.Name, l.Len)
	return e
}

func FullChrlens() {
	fa := ParseFasta(os.Stdin)
	lens := Chrlens(fa)
	for l, e := range lens {
		if e != nil {
			log.Fatal(e)
		}
		if e := PrintFaLen(os.Stdout, l); e != nil {
			log.Fatal(e)
		}
	}
}
