package fastats

import (
	"os"
	"iter"
	"io"
	"fmt"
	"strings"

	"github.com/jgbaldwinbrown/iterh"
	"github.com/jgbaldwinbrown/csvh"
)

type FaiEntry struct {
	Name string
	Length int64
	Offset int64
	Linebases int64
	Linewidth int64
	Qualoffset int64
}

func ParseFai(r io.Reader) iter.Seq2[FaiEntry, error] {
	return func(yield func(FaiEntry, error) bool) {
		cr := csvh.CsvIn(r)
		for l, e := cr.Read(); e != io.EOF; l, e = cr.Read() {
			if e != nil {
				if !yield(FaiEntry{}, e) {
					return
				}
			}
			var f FaiEntry
			var e error
			if len(l) == 5 {
				_, e = csvh.Scan(l, &f.Name, &f.Length, &f.Offset, &f.Linebases, &f.Linewidth)
				f.Qualoffset = -1
			} else if len(l) == 6 {
				_, e = csvh.Scan(l, &f.Name, &f.Length, &f.Offset, &f.Linebases, &f.Linewidth, &f.Qualoffset)
			}
			if !yield(f, e) {
				return
			}
		}
	}
}

func FaiMap(it iter.Seq[FaiEntry]) map[string]FaiEntry {
	m := map[string]FaiEntry{}
	for f := range it {
		m[f.Name] = f
	}
	return m
}

func FastaFaiCoords(f FaiEntry, start int64, end int64) iter.Seq[Span] {
	return func(yield func(Span) bool) {
		lend := f.Linewidth - f.Linebases
		rawstart := start + ((start / f.Linebases) * lend) + f.Offset
		rawend := end + ((end / f.Linebases) * lend) + f.Offset
		pos := rawstart
		if pos % f.Linewidth != 0 {
			edge := (((pos - f.Offset) / f.Linewidth) * f.Linewidth) + f.Offset + f.Linebases
			if !yield(Span{pos, edge}) {
				return
			}
			pos = edge + lend
		}
		for ; pos < rawend; pos += f.Linewidth {
			if !yield(Span{pos, pos + f.Linebases}) {
				return
			}
		}
	}
}

// func fastqFaiCoords(f FaiEntry, start int64, end int64) iter.Seq2[Span, Span] {
// 	panic(errors.New("fastqFaiCoords not implemented!"))
// 	return nil
// }

type FaiHandle struct {
	Fai map[string]FaiEntry
	File io.ReadSeekCloser
}

func (h *FaiHandle) FastaFaiExtract(chr string, start, end int64) (FaEntry, error) {
	ent, ok := h.Fai[chr]
	if !ok {
		return FaEntry{}, fmt.Errorf("FastaFaiExtract: no chromosome %v in Fai", chr)
	}

	filesize, e := h.File.Seek(0, io.SeekEnd)
	if e != nil {
		return FaEntry{}, e
	}

	spans := FastaFaiCoords(ent, start, end)
	var seq strings.Builder
	for span := range spans {
		if span.End > filesize {
			return FaEntry{}, fmt.Errorf("FastaFaiExtract: tried to extract span %v past end of file %v", span, filesize)
		}
		_, e := h.File.Seek(span.Start, io.SeekStart)
		if e != nil {
			return FaEntry{}, e
		}
		_, e = io.CopyN(&seq, h.File, span.End - span.Start)
		if e != nil {
			return FaEntry{}, e
		}
	}
	return FaEntry{Header: chr, Seq: seq.String()}, nil
}

func (h *FaiHandle) Close() error {
	return h.File.Close()
}

func NewFaiHandle(faqPath, faiPath string) (*FaiHandle, error) {
	faq, e := os.Open(faqPath)
	if e != nil {
		return nil, e
	}

	fai := iterh.PathIter(faiPath, ParseFai)
	fai2 := iterh.BreakOnError(fai, &e)
	faiMap := FaiMap(fai2)
	if e != nil {
		return nil, e
	}
	return &FaiHandle{Fai: faiMap, File: faq}, nil
}
