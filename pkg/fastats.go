package fastats

import (
	"encoding/json"
	"golang.org/x/exp/slices"
	"iter"
	"math"
	"os"
)

type FaStats struct {
	Bp      int64
	MeanLen float64
	NumSeqs int64
	N50     int64
	L50     int64
	BpInN50 int64
	N90     int64
	L90     int64
	BpInN90 int64
}

func RevsortedLens(it iter.Seq2[FaLen, error]) ([]FaLen, error) {
	out, err := CollectErr(it)
	if err != nil {
		return nil, err
	}

	slices.SortFunc(out, func(i, j FaLen) int {
		return int(j.Len - i.Len)
	})

	return out, nil
}

func BasicStats(lens []FaLen, s *FaStats) {
	s.Bp = 0
	for _, l := range lens {
		s.Bp += l.Len
	}
	s.NumSeqs = int64(len(lens))
	s.MeanLen = float64(s.Bp) / float64(s.NumSeqs)
}

func NStats(lens []FaLen, totalbp int64, nfrac float64, n, l, bpInN *int64) {
	totfrac := int64(math.Ceil(float64(totalbp) * nfrac))
	*bpInN = 0
	*l = 0
	for _, falen := range lens {
		(*l)++
		(*bpInN) += falen.Len
		*n = falen.Len
		if *bpInN >= totfrac {
			break
		}
	}
}

func N50Stats(lens []FaLen, s *FaStats) {
	NStats(lens, s.Bp, 0.5, &s.N50, &s.L50, &s.BpInN50)
}

func N90Stats(lens []FaLen, s *FaStats) {
	NStats(lens, s.Bp, 0.9, &s.N90, &s.L90, &s.BpInN90)
}

func Stats(it iter.Seq2[FaLen, error]) (FaStats, error) {
	lens, err := RevsortedLens(it)
	if err != nil {
		return FaStats{}, err
	}

	var s FaStats
	BasicStats(lens, &s)
	N50Stats(lens, &s)
	N90Stats(lens, &s)
	return s, nil
}

func FullFaStats() {
	stats, err := Stats(Chrlens(ParseFasta(os.Stdin)))
	if err != nil {
		panic(err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "\t")
	err = enc.Encode(stats)
	if err != nil {
		panic(err)
	}
}
