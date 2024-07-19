package fastats

import (
	"iter"
	"strings"
)

func toFaEntry[F FaEnter](f F) FaEntry {
	return FaEntry{Header: f.FaHeader(), Seq: f.FaSeq()}
}

func FaQualMerge[F FaEnter](fit iter.Seq2[F, error], qit iter.Seq2[[]int64, error]) iter.Seq2[FqEntry, error] {
	qp, cancel := iter.Pull2(qit)
	defer cancel()
	var b strings.Builder
	return func(yield func(FqEntry, error) bool) {
		for fa, e := range fit {
			if e != nil {
				if !yield(FqEntry{}, e) {
					return
				}
			}

			var q []int64
			var ok bool
			q, e, ok = qp()
			if !ok {
				return
			}
			if e != nil {
				if !yield(FqEntry{}, e) {
					return
				}
			}

			b.Reset()
			e = AppendQualsToAscii(&b, q)

			if !yield(FqEntry{FaEntry: toFaEntry(fa), Qual: b.String()}, e) {
				return
			}
		}
	}
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

func FaToFq[FA FaEnter](fit iter.Seq2[FA, error], qual byte) iter.Seq2[FqEntry, error] {
	base := BuildQual(qual, 1024)
	return func(yield func(FqEntry, error) bool) {
		for f, e := range fit {
			if e != nil {
				if !yield(FqEntry{}, e) {
					return
				}
			}
			if len(f.FaSeq()) <= 1024 {
				if !yield(FqEntry{FaEntry: toFaEntry(f), Qual: base[:len(f.FaSeq())]}, nil) {
					return
				}
			} else {
				if !yield(FqEntry{FaEntry: toFaEntry(f), Qual: BuildQual(qual, len(f.FaSeq()))}, nil) {
					return
				}
			}
		}
	}
}
