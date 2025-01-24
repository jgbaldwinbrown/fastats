package fastats

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"iter"
	"log"
	"os"

	"github.com/jgbaldwinbrown/zfile"
	"github.com/jgbaldwinbrown/iterh"
)

func BedZip[B1 BedEnter[[]string], B2 BedEnter[[]string]](it1 iter.Seq[B1], it2 iter.Seq[B2]) []BedEntry[[]string] {
	matched := ZipMatches(it1, it2)
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
	Bed1    string
	Bed2    string
}

func PassLine(r io.ByteReader, w io.ByteWriter) error {
	for {
		val, e := r.ReadByte()
		if e == io.EOF {
			break
		}
		if e != nil {
			return e
		}
		e = w.WriteByte(val)
		if e != nil {
			return e
		}
		if val == '\n' {
			break
		}
	}
	return nil
}

type byteReaderReader interface {
	io.Reader
	io.ByteReader
}

func FullBedZip() {
	var f bedZipFlags
	flag.BoolVar(&f.Header1, "h1", false, "first bed file has a header")
	flag.BoolVar(&f.Header2, "h2", false, "second bed file has a header")
	flag.StringVar(&f.Bed1, "b1", "", "first bed file")
	flag.StringVar(&f.Bed2, "b2", "", "second bed file")
	flag.Parse()

	var err error
	r1, e := zfile.Open(f.Bed1)
	if e != nil {
		log.Fatal(e)
	}
	defer r1.Close()
	br1 := bufio.NewReader(r1)
	r2, e := zfile.Open(f.Bed2)
	if e != nil {
		log.Fatal(e)
	}
	defer r2.Close()
	br2 := bufio.NewReader(r2)
	w := bufio.NewWriter(os.Stdout)
	defer func() {
		e := w.Flush()
		if e != nil {
			log.Fatal(e)
		}
	}()

	if f.Header1 {
		e := PassLine(br1, w)
		if e != nil {
			log.Fatal(e)
		}
	}
	if f.Header2 {
		e := PassLine(br2, w)
		if e != nil {
			log.Fatal(e)
		}
	}

	b1 := iterh.BreakOnError(ParseBedFlat(br1), &err)
	b2 := iterh.BreakOnError(ParseBedFlat(br2), &err)

	out := BedZip(b1, b2)
	if err != nil {
		log.Fatal(err)
	}

	for _, b := range out {
		fmt.Fprintf(w, "%v\t%v\t%v", b.Chr, b.Start, b.End)
		for _, f := range b.Fields {
			fmt.Fprintf(w, "\t%v", f)
		}
		fmt.Fprintf(w, "\n")
	}
}
