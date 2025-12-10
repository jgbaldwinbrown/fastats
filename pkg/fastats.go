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

func RevsortedLensEfficient(it iter.Seq2[FaLen, error]) (lens []int64, counts map[int64]int64, err error) {
	counts = map[int64]int64{}
	for falen, e := range it {
		if e != nil {
			return nil, nil, e
		}
		counts[falen.Len]++
	}
	lens = make([]int64, 0, len(counts))
	for length, _ := range counts {
		lens = append(lens, length)
	}
	slices.SortFunc(lens, func(i, j int64) int {
		return int(j - i)
	})
	return lens, counts, nil
	
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

func BasicStatsEfficient(lens []int64, counts map[int64]int64, s *FaStats) {
	s.Bp = 0
	s.NumSeqs = 0
	for _, l := range lens {
		count := counts[l]
		s.Bp += l * count
		s.NumSeqs += count
	}
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

func NStatsEfficient(lens []int64, counts map[int64]int64, totalbp int64, nfrac float64, n, l, bpInN *int64) {
	totfrac := int64(math.Ceil(float64(totalbp) * nfrac))
	*bpInN = 0
	*l = 0
	for _, length := range lens {
		count := counts[length]
		var i int64
		for i = 0; i < count; i++ {
			(*l)++
			(*bpInN) += length
			*n = length
			if *bpInN >= totfrac {
				return
			}
		}
	}
}

func N50Stats(lens []FaLen, s *FaStats) {
	NStats(lens, s.Bp, 0.5, &s.N50, &s.L50, &s.BpInN50)
}

func N50StatsEfficient(lens []int64, count map[int64]int64, s *FaStats) {
	NStatsEfficient(lens, count, s.Bp, 0.5, &s.N50, &s.L50, &s.BpInN50)
}

func N90Stats(lens []FaLen, s *FaStats) {
	NStats(lens, s.Bp, 0.9, &s.N90, &s.L90, &s.BpInN90)
}

func N90StatsEfficient(lens []int64, count map[int64]int64, s *FaStats) {
	NStatsEfficient(lens, count, s.Bp, 0.9, &s.N90, &s.L90, &s.BpInN90)
}

func Stats(it iter.Seq2[FaLen, error]) (FaStats, error) {
	lens, counts, err := RevsortedLensEfficient(it)
	if err != nil {
		return FaStats{}, err
	}

	var s FaStats
	BasicStatsEfficient(lens, counts, &s)
	N50StatsEfficient(lens, counts, &s)
	N90StatsEfficient(lens, counts, &s)
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
