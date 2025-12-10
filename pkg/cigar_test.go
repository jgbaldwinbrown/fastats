package fastats

import (
	"reflect"
	"testing"
)

func TestParseCIGAR(t *testing.T) {
	in := "33M22M45D3I2S"
	exp := []CIGAREntry{
		CIGAREntry{'M', 33},
		CIGAREntry{'M', 22},
		CIGAREntry{'D', 45},
		CIGAREntry{'I', 3},
		CIGAREntry{'S', 2},
	}
	got, e := ParseCIGAR(in)
	if e != nil {
		t.Error(e)
	}
	if !reflect.DeepEqual(got, exp) {
		t.Errorf("got %v != exp %v", got, exp)
	}
}
