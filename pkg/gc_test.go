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
	for win, e := range gc {
		if e != nil {
			panic(e)
		}
		fmt.Println(win)
	}
}

func TestGCPerWin(t *testing.T) {
	in := FaWins(ParseFasta(strings.NewReader(infa)), 3, 2)
	for win, e := range in {
		if e != nil {
			panic(e)
		}
		fmt.Println(win)
		fmt.Println(GCFrac(win.Fields))
	}
}
