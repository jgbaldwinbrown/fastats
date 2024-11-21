package fastats

import (
	"fmt"
	"testing"
	"strings"
	"slices"

	"github.com/jgbaldwinbrown/iterh"
)

const fastaEx = `>1
aatctatgtcagtacagcgttgcgggtact`

const gffEx = `1	fake	mRNA	0	20	.	+	.	ID=transcript1
1	fake	CDS	5	15	.	+	.	ID=cds1;Parent=transcript1`

func TestParseGffFlat(t *testing.T) {
	var e error
	gff := slices.Collect(iterh.BreakOnError(ParseGffFlat(strings.NewReader(gffEx)), &e))

	fmt.Printf("gff: %#v; err: %#v\n", gff, e)
}
