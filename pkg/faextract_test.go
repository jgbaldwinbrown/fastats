package fastats

import (
	"fmt"
	"strings"
	"testing"
)

const exampleFa = `>1
gtcagtgtacgtac
gactgta
>2
gtcagtca
`

const exampleBed = `1	3	5
2	2	4
`

func TestExtractFasta(t *testing.T) {
	fa := ParseFasta(strings.NewReader(exampleFa))
	bed := ParseBedFlat(strings.NewReader(exampleBed))
	for fa, err := range ExtractFasta(fa, bed) {
		if err != nil {
			t.Error(err)
		}
		fmt.Println(fa)
	}
}

func TestGcChrSpanFa(t *testing.T) {
	fa := ParseFasta(strings.NewReader(exampleFa))
	bed := ParseBedFlat(strings.NewReader(exampleBed))
	csf := ExtractChrSpanFa(fa, bed)
	cs := GcChrSpanFa(csf)
	for stat, err := range cs {
		if err != nil {
			t.Error(err)
		}
		fmt.Println(stat)
	}
}
