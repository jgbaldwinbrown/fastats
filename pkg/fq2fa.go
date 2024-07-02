package fastats

import (
	"os"
	"bufio"
	"fmt"
	"log"
)

func FullToFa() {
	it := ParseFastq(os.Stdin)
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	for f, e := range it {
		if e != nil {
			log.Fatal(e)
		}
		_, e := fmt.Fprintf(w, ">%v\n%v\n", f.FaHeader(), f.FaSeq())
		if e != nil {
			panic(e)
		}
	}
}
