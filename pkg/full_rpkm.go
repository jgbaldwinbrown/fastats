package fastats

import (
	"os"
	"github.com/jgbaldwinbrown/iterh"
	"fmt"
	"log"
	"bufio"
	"flag"
	"iter"
)

type FullRpkmFlags struct {
	BedPath string
	TotalCoverage float64
}

func FullRpkm() {
	var f FullRpkmFlags
	flag.StringVar(&f.BedPath, "bed", "", "Input bed file (required).")
	flag.Float64Var(&f.TotalCoverage, "cov", -1.0, "Total coverage to substitute for the one calculated by the program.")
	flag.Parse()

	if f.BedPath == "" {
		log.Fatal(fmt.Errorf("Missing -bed option"))
	}
	cov1, errp1 := iterh.BreakWithError(iterh.PathIter(f.BedPath, ParseBedGraph))
	scov1 := SpreadBed(cov1)

	var rpkm1 iter.Seq[BedEntry[float64]]
	if f.TotalCoverage < 0.0 {
		rpkm1, _ = RpkmAndTotal(scov1)
	} else {
		rpkm1 = Rpkm(scov1, f.TotalCoverage)
	}

	w := bufio.NewWriter(os.Stdout)
	defer func() {
		e := w.Flush()
		if e != nil {
			log.Fatal(e)
		}
	}()

	for b := range rpkm1 {
		_, e := fmt.Fprintf(w, "%v\t%v\t%v\t%v\n", b.Chr, b.Start, b.End, b.Fields)
		if e != nil {
			log.Fatal(e)
		}
	}
	if *errp1 != nil {
		log.Fatal(*errp1)
	}
}
