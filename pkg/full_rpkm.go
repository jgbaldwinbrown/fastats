package fastats

import (
	"os"
	"github.com/jgbaldwinbrown/iterh"
	"fmt"
	"log"
	"bufio"
)

func FullRpkm() {
	if len(os.Args) < 2 {
		fmt.Printf("usage: %v bed.bed \n", os.Args[0])
		log.Fatal(fmt.Errorf("Not enough args: %v", os.Args))
	}
	cov1, errp1 := iterh.BreakWithError(iterh.PathIter(os.Args[1], ParseBedGraph))
	scov1 := SpreadBed(cov1)
	rpkm1, _ := RpkmAndTotal(scov1)

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
