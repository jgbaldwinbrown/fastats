package fastats

import (
	"testing"
	"fmt"
	"strings"
)

const infa = `>1
acacgtgt

>2
ttggttggttggttgg

>3
aaaaaaaaaaaaaaaa
`

func TestGC(t *testing.T) {
	in := FaWins(ParseFasta(strings.NewReader(infa)), 3, 2)
	gc := GCIter(in)
	e := gc.Iterate(func(win BedEntry[float64]) error {
		fmt.Println(win)
		return nil
	})
	if e != nil {
		panic(e)
	}
}

func TestGCPerWin(t *testing.T) {
	in := FaWins(ParseFasta(strings.NewReader(infa)), 3, 2)
	e := in.Iterate(func(win BedEntry[string]) error {
		fmt.Println(win)
		fmt.Println(GCFrac(win.Fields))
		return nil
	})
	if e != nil {
		panic(e)
	}
}
