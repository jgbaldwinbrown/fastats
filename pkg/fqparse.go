package fastats

import (
	"log"
	"encoding/json"
	"os"
	"fmt"
	"bufio"
	"io"
	"iter"
)

type FqEntry struct {
	FaEntry
	Qual string
}

func (f FqEntry) FqQual() string { return f.Qual }

type FqEnter interface {
	FaEnter
	FqQual() string
}

func ToFqEntry[F FqEnter](f F) FqEntry {
	return FqEntry {
		FaEntry: ToFaEntry(f),
		Qual: f.FqQual(),
	}
}

func ScanFour(dest []string, s *bufio.Scanner) ([]string, error) {
	dest = dest[:0]
	for i := 0; i < 4; i++ {
		notDone := s.Scan()
		if !notDone {
			return dest, io.EOF
		}
		if s.Err() != nil {
			return dest, s.Err()
		}
		dest = append(dest, s.Text())
	}
	return dest, nil
}

func parseFastq(r io.Reader, yield func(FqEntry, error) bool) {
	s := bufio.NewScanner(r)
	s.Buffer([]byte{}, 1e12)
	var lines []string
	for lines, err := ScanFour(lines, s); err != io.EOF; lines, err = ScanFour(lines, s) {
		if err != nil {
			if !yield(FqEntry{}, err) {
				return
			}
		}
		if len(lines[0]) < 1 {
			if !yield(FqEntry{}, fmt.Errorf("parseFastq: empty header line")) {
				return
			}
		}
		if !yield(FqEntry{FaEntry: FaEntry{Header: lines[0][1:], Seq: lines[1]}, Qual: lines[3]}, nil) {
			return
		}
	}
}

func ParseFastq(r io.Reader) iter.Seq2[FqEntry, error] {
	return func(yield func(FqEntry, error) bool) {
		parseFastq(r, yield)
	}
}

func FqChrlens[F FqEnter](it iter.Seq2[F, error]) iter.Seq2[FaLen, error] {
	return func(yield func(FaLen, error) bool) {
		for f, e := range it {
			falen := Chrlen(f)
			if !yield(falen, e) {
				return
			}
		}
	}
}

func FullFqChrlens() {
	lens := FqChrlens(ParseFastq(os.Stdin))

	for l, e := range lens {
		if e != nil {
			log.Fatal(e)
		}
		if e := PrintFaLen(os.Stdout, l); e != nil {
			log.Fatal(e)
		}
	}
}

func FullFqStats() {
	stats, err := Stats(FqChrlens(ParseFastq(os.Stdin)))
	if err != nil {
		panic(err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "\t")
	err = enc.Encode(stats)
	if err != nil {
		panic(err)
	}
}
