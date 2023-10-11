package fastats

import (
	"encoding/json"
	"os"
	"fmt"
	"bufio"
	"io"
)

type FqEntry struct {
	FaEntry
	Qual string
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

func parseFastq(r io.Reader, yield func(FqEntry) error) error {
	s := bufio.NewScanner(r)
	s.Buffer([]byte{}, 1e12)
	var lines []string
	for lines, err := ScanFour(lines, s); err != io.EOF; lines, err = ScanFour(lines, s) {
		if err != nil {
			return err
		}
		if len(lines[0]) < 1 {
			return fmt.Errorf("parseFastq: empty header line")
		}
		err = yield(FqEntry{FaEntry: FaEntry{Header: lines[0][1:], Seq: lines[1]}, Qual: lines[3]})
		if err != nil {
			return err
		}
	}
	return nil
}

func ParseFastq(r io.Reader) *Iterator[FqEntry] {
	return &Iterator[FqEntry]{Iteratef: func(yield func(FqEntry) error) error {
		return parseFastq(r, yield)
	}}
}

func FqChrlens(it Iter[FqEntry]) *Iterator[FaLen] {
	return &Iterator[FaLen]{Iteratef: func(yield func(FaLen) error) error {
		return it.Iterate(func (f FqEntry) error {
			falen := Chrlen(f.FaEntry)
			return yield(falen)
		})
	}}
}

func FullFqChrlens() {
	lens := FqChrlens(ParseFastq(os.Stdin))

	err := lens.Iterate(func(l FaLen) error {
		return PrintFaLen(os.Stdout, l)
	})
	if err != nil {
		panic(err)
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
