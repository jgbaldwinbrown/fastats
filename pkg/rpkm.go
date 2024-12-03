package fastats

import (
	"iter"
)

func BedGraphSum[B BedEnter[float64]](it iter.Seq[B]) float64 {
	sum := 0.0
	for b := range it {
		sum += b.BedFields()
	}
	return sum
}

func RpkmOne[B BedEnter[float64]](b B, totalCov float64) float64 {
	length := float64(b.SpanEnd() - b.SpanStart())
	cov := b.BedFields()
	covper1kb := cov / (length / 1000)
	return covper1kb / (totalCov / 1e6)
}

func Rpkm[B BedEnter[float64]](it iter.Seq[B], totalCov float64) iter.Seq[BedEntry[float64]] {
	return func(y func(BedEntry[float64]) bool) {
		for b := range it {
			out := ToBedEntry(b)
			out.Fields = RpkmOne(b, totalCov)
			if !y(out) {
				return
			}
		}
	}
}

// "it" must be reuseable
func RpkmAndTotal[B BedEnter[float64]](it iter.Seq[B]) (rpkm iter.Seq[BedEntry[float64]], totalCov float64) {
	totalCov = BedGraphSum(it)
	return Rpkm(it, totalCov), totalCov
}
