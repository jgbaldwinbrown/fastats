package fastats

import (
	"testing"
	"strings"
	"fmt"
)

const testNwk = "(,A,B:5,:6)C;"

func TestParseNewick(t *testing.T) {
	n, e := ParseNewick(strings.NewReader(testNwk))
	if e != nil {
		t.Error(e)
	}
	fmt.Println(n)
}
