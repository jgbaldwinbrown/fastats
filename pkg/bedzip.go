package fastats

import (
	"os"
	"iter"
	"fmt"
	"log"
	"bufio"

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

func FullBedZip() {
	args := os.Args
	if len(args) < 3 {
		log.Printf("usage: %s bed1.bed bed2.bed\n", args[0])
		log.Fatal(fmt.Errorf("missing arguments"))
	}
	var err error
	b1 := iterh.BreakOnError(iterh.PathIter(args[1], ParseBedFlat), &err)
	b2 := iterh.BreakOnError(iterh.PathIter(args[2], ParseBedFlat), &err)

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
