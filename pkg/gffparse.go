package fastats

import (
	"strconv"
	"encoding/csv"
	"fmt"
	"io"
	"iter"
	"regexp"
)

type GffHead struct {
	ChrSpan
	Source   string
	Type     string
	Score    float64
	HasScore bool
	Strand   byte
	Phase    int
	HasPhase bool
}

type GffHeader interface {
	ChrSpanner
	GffSource() string
	GffType() string
	GffScore() float64
	GffHasScore() bool
	GffStrand() byte
	GffPhase() int
	GffHasPhase() bool
}

func (g GffHead) GffSource() string { return g.Source }
func (g GffHead) GffType() string   { return g.Type }
func (g GffHead) GffScore() float64 { return g.Score }
func (g GffHead) GffHasScore() bool { return g.HasScore }
func (g GffHead) GffStrand() byte   { return g.Strand }
func (g GffHead) GffPhase() int     { return g.Phase }
func (g GffHead) GffHasPhase() bool { return g.HasPhase }

func ToGffHead[G GffHeader](g G) GffHead {
	if ptr, ok := any(&g).(*GffHead); ok {
		return *ptr
	}
	return GffHead{
		ChrSpan:  ToChrSpan(g),
		Source:   g.GffSource(),
		Type:     g.GffType(),
		Score:    g.GffScore(),
		HasScore: g.GffHasScore(),
		Strand:   g.GffStrand(),
		Phase:    g.GffPhase(),
		HasPhase: g.GffHasPhase(),
	}
}

type GffEntry[AttT any] struct {
	GffHead
	Attributes AttT
}

type GffEnter[AttT any] interface {
	GffHeader
	GffAttributes() AttT
}

func (g GffEntry[T]) GffAttributes() T {
	return g.Attributes
}

func ToGffEntry[G GffEnter[AttT], AttT any](g G) GffEntry[AttT] {
	if ptr, ok := any(&g).(*GffEntry[AttT]); ok {
		return *ptr
	}
	return GffEntry[AttT]{
		GffHead:    ToGffHead(g),
		Attributes: g.GffAttributes(),
	}
}

func ParseGffEntry[AT any](line []string, attributeParse func(string) (AT, error)) (GffEntry[AT], error) {
	var g GffEntry[AT]
	if len(line) < 8 {
		return g, fmt.Errorf("ParseBedEntry: len(line) %v < 8", len(line))
	}

	var e error
	g.Chr = line[0]
	g.Source = line[1]
	g.Type = line[2]
	g.Start, e = strconv.ParseInt(line[3], 0, 64)
	if e != nil {
		return g, fmt.Errorf("ParseGffEntry: Start: %w", e)
	}
	g.Start--
	g.End, e = strconv.ParseInt(line[4], 0, 64)
	if e != nil {
		return g, fmt.Errorf("ParseGffEntry: End: %w", e)
	}

	if line[5] != "." {
		g.HasScore = true
		g.Score, e = strconv.ParseFloat(line[5], 64)
		if e != nil {
			return g, fmt.Errorf("ParseGffEntry: Score: %w", e)
		}
	}

	if len(line[6]) > 0 {
		g.Strand = line[6][0]
	}

	if line[7] != "." {
		g.HasPhase = true
		tempPhase, e := strconv.ParseInt(line[7], 0, 64)
		if  e != nil {
			return g, fmt.Errorf("ParseGffEntry: Phase: %w", e)
		}
		g.Phase = int(tempPhase)
	}

	g.Attributes, e = attributeParse(line[8])
	if e != nil {
		return g, fmt.Errorf("ParseGffEntry: Attributes: %w", e)
	}
	return g, nil
}

func ParseGff[AT any](r io.Reader, attributeParse func(string) (AT, error)) iter.Seq2[GffEntry[AT], error] {
	return func(yield func(GffEntry[AT], error) bool) {
		cr := csv.NewReader(r)
		cr.LazyQuotes = true
		cr.Comma = rune('\t')
		cr.ReuseRecord = true
		cr.FieldsPerRecord = -1

		for l, e := cr.Read(); e != io.EOF; l, e = cr.Read() {
			b, e := ParseGffEntry(l, attributeParse)
			if !yield(b, e) {
				return
			}
		}
	}
}

type AttributePair struct {
	Tag   string
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

func ParseGffFlat(r io.Reader) iter.Seq2[GffEntry[[]AttributePair], error] {
	return ParseGff(r, ParseAttributePairs)
}
