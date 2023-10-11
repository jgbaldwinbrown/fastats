package fastats

import (
	"regexp"
	"fmt"
	"io"
	"encoding/csv"
)

type GffEntry[AttT any] struct {
	ChrSpan
	Source string
	Type string
	Score float64
	HasScore bool
	Strand byte
	Phase int
	HasPhase bool
	Attributes AttT
}

func ParseGffEntry[AT any](line []string, attributeParse func(string) (AT, error)) (GffEntry[AT], error) {
	var g GffEntry[AT]
	if len(line) < 8 {
		return g, fmt.Errorf("ParseBedEntry: len(line) %v < 8", len(line))
	}

	var scoreStr string
	var phaseStr string
	_, e := Scan(line[:8], &g.Chr, &g.Source, &g.Type, &g.Start, &g.End, &scoreStr, &g.Strand, &phaseStr)
	if e != nil {
		return g, e
	}

	if scoreStr != "." {
		g.HasScore = true
		_, e := fmt.Sscanf(scoreStr, "%v", &g.Score)
		if e != nil {
			return g, e
		}
	}

	if phaseStr != "." {
		g.HasScore = true
		_, e := fmt.Sscanf(phaseStr, "%v", &g.Phase)
		if e != nil {
			return g, e
		}
	}

	g.Attributes, e = attributeParse(line[8])
	return g, e
}

func ParseGff[AT any](r io.Reader, attributeParse func(string) (AT, error)) *Iterator[GffEntry[AT]] {
	return &Iterator[GffEntry[AT]]{Iteratef: func(yield func(GffEntry[AT]) error) error {
		cr := csv.NewReader(r)
		cr.LazyQuotes = true
		cr.Comma = rune('\t')
		cr.ReuseRecord = true
		cr.FieldsPerRecord = -1

		for l, e := cr.Read(); e != io.EOF; l, e = cr.Read() {
			b, e := ParseGffEntry(l, attributeParse)
			if e != nil {
				return e
			}
			e = yield(b)
			if e != nil {
				return e
			}
		}

		return nil
	}}
}

type AttributePair struct {
	Tag string
	Value string
}

var attRe = regexp.MustCompile(`([^=]*)=([^;]*);?`)

func ParseAttributePairs(field string) ([]AttributePair, error) {
	matches := attRe.FindAllStringSubmatch(field, -1)
	out := make([]AttributePair, 0, len(matches))
	for _, match := range matches {
		out = append(out, AttributePair{match[1], match[2]})
	}
	return out, nil
}

func ParseGffFlat(r io.Reader) *Iterator[GffEntry[[]AttributePair]] {
	return ParseGff(r, ParseAttributePairs)
}
