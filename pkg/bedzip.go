package fastats

import (
	"os"
	"iter"
	"fmt"
	"log"
	"bufio"
	"flag"
	"io"

	"github.com/jgbaldwinbrown/iterh"
)

func BedZip[B1 BedEnter[[]string], B2 BedEnter[[]string]](it1 iter.Seq[B1], it2 iter.Seq[B2]) []BedEntry[[]string] {
	matched := iterh.ZipMatches(it1, it2)
	out := make([]BedEntry[[]string], 0, len(matched))
	for chrspan, tup := range matched {
		if tup.Ok {
			out = append(out, BedEntry[[]string]{ChrSpan: chrspan, Fields: append(tup.V1, tup.V2...)})
		}
		delete(matched, chrspan)
	}
	SortBed(out)
	return out
}

type bedZipFlags struct {
	Header1 bool
	Header2 bool
	Bed1 string
	Bed2 string
}

func ChompLine(r io.ByteReader) error {
	for {
		val, e := r.ReadByte()
		if e == io.EOF {
			break
		}
		if e != nil {
			return e
		}
		if val == '\n' {
			break
		}
	}
	return nil
}

func ParseBedFlatHeader(r io.Reader) iter.Seq2[BedEntry[[]string], error] {
	return func(y func(BedEntry[[]string], error) bool) {
		_, e := ChompLine(r)
		if e != nil && !y(BedEntry[[]string]{}, e) {
			return
		}
		for b, err := range ParseBedFlat(r) {
			if !y(b, err) {
				return
			}
		}
	}
}

func FullBedZip() {
	var f bedZipFlags
	flag.BoolVar(&f.Header1, "h1", false, "first bed file has a header")
	flag.BoolVar(&f.Header2, "h2", false, "second bed file has a header")
	flag.StringVar(&f.Bed1, "h1", "", "first bed file")
	flag.StringVar(&f.Bed2, "h2", "", "second bed file")
	flag.Parse()

	var err error
	var b1, b2 iter.Seq2[BedEntry[[]string], error]
	if (f.Header1) {
		b1 := iterh.BreakOnError(iterh.PathIter(f.Bed1, ParseBedFlatHeader), &err)
	} else {
		b1 := iterh.BreakOnError(iterh.PathIter(f.Bed1, ParseBedFlat), &err)
	}
	if (f.Header2) {
		b2 := iterh.BreakOnError(iterh.PathIter(f.Bed2, ParseBedFlatHeader), &err)
	} else {
		b2 := iterh.BreakOnError(iterh.PathIter(f.Bed2, ParseBedFlat), &err)
	}

	out := BedZip(b1, b2)
	if err != nil {
		log.Fatal(err)
	}

	w := bufio.NewWriter(os.Stdout)
	defer func() {
		e := w.Flush()
		if e != nil {
			log.Fatal(e)
		}
	}()
	for _, b := range out {
		fmt.Fprintf(w, "%v\t%v\t%v", b.Chr, b.Start, b.End)
		for _, f := range b.Fields {
			fmt.Fprintf(w, "\t%v", f)
		}
		fmt.Fprintf(w, "\n")
	}
}
