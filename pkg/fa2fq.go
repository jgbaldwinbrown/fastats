package fastats

import (
	"github.com/jgbaldwinbrown/iter"
	"io"
	"strings"
)

func FaQualMerge(fit iter.Puller[FaEntry], qit iter.Puller[[]int64]) iter.Iter[FqEntry] {
	var b strings.Builder
	return &iter.Iterator[FqEntry]{Iteratef: func(yield func(FqEntry) error) error {
		for fa, e := fit.Next(); e != io.EOF; fa, e = fit.Next() {
			if e != nil {
				return e
			}

			var q []int64
			q, e = qit.Next()
			if e != nil {
				return e
			}

			b.Reset()
			if e = AppendQualsToAscii(&b, q); e != nil {
				return e
			}

			e = yield(FqEntry{FaEntry: fa, Qual: b.String()})
			if e != nil {
				return e
			}
		}
		return nil
	}}
}

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func BuildQual(qual byte, length int) string {
	var b strings.Builder
	for i := 0; i < length; i++ {
		Must(b.WriteByte(qual))
	}
	return b.String()
}

func FaToFq(fit iter.Iter[FaEntry], qual byte) *iter.Iterator[FqEntry] {
	base := BuildQual(qual, 1024)
	return &iter.Iterator[FqEntry]{Iteratef: func(yield func(FqEntry) error) error {
		return fit.Iterate(func(f FaEntry) error {
			var err error
			if len(f.Seq) <= 1024 {
				err = yield(FqEntry{FaEntry: f, Qual: base[:len(f.Seq)]})
			} else {
				err = yield(FqEntry{FaEntry: f, Qual: BuildQual(qual, len(f.Seq))})
			}
			return err
		})
	}}
}
