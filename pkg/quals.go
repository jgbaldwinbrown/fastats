package fastats

import (
	"bufio"
	"encoding/json"
	"golang.org/x/exp/slices"
	"io"
	"iter"
	"math"
	"os"
)

// q = -10 log_10(p)
// p = 10^(-q/10)

func QualToAscii(qual int64) byte {
	return byte(qual + 33)
}

func AppendQualsToAscii(w io.Writer, quals []int64) error {
	var bw io.ByteWriter
	if wbw, ok := w.(io.ByteWriter); ok {
		bw = wbw
	} else {
		b := bufio.NewWriter(w)
		defer b.Flush()
		bw = b
	}

	for _, qual := range quals {
		e := bw.WriteByte(QualToAscii(qual))
		if e != nil {
			return e
		}
	}
	return nil
}

func QualScore(qual byte) float64 {
	fqual := float64(qual - 33)
	p := math.Pow(10, (-fqual)/10)
	return p
}

func ScoreQual(score float64) byte {
	fqual := -10 * math.Log10(score)
	qual := 33 + byte(math.Round(fqual))
	return qual
}

func AppendQualScores(dest []float64, quals string) []float64 {
	for _, qual := range []byte(quals) {
		dest = append(dest, QualScore(qual))
	}
	return dest
}

func AppendScoreQuals(w io.Writer, scores []float64) error {
	var bw io.ByteWriter
	if wbw, ok := w.(io.ByteWriter); ok {
		bw = wbw
	} else {
		b := bufio.NewWriter(w)
		defer b.Flush()
		bw = b
	}

	for _, score := range scores {
		e := bw.WriteByte(ScoreQual(score))
		if e != nil {
			return e
		}
	}
	return nil
}

func MeanQual[F FqEnter](it iter.Seq2[F, error]) (float64, error) {
	sum := 0.0
	count := 0.0
	var scores []float64
	for f, err := range it {
		if err != nil {
			return 0, err
		}
		scores = AppendQualScores(scores[:0], f.FqQual())
		for _, score := range scores {
			sum += score
		}
		count += float64(len(scores))
	}
	return sum / count, nil
}

func Mean(fs ...float64) float64 {
	sum := 0.0
	count := 0.0
	for _, f := range fs {
		sum += f
		count++
	}
	return sum / count
}

func MeanReadQual[F FqEnter](it iter.Seq2[F, error]) (float64, error) {
	sum := 0.0
	count := 0.0
	var scores []float64
	for f, err := range it {
		if err != nil {
			return 0, err
		}
		scores = AppendQualScores(scores[:0], f.FqQual())
		sum += Mean(scores...)
		count++
	}
	return sum / count, nil
}

func GrowLen[T any](s []T, n int) []T {
	s = slices.Grow(s, n)
	if len(s) < n {
		s = s[:n]
	}
	return s
}

func QualPerPos[F FqEnter](it iter.Seq2[F, error]) ([]float64, error) {
	var sum []float64
	var count []float64
	var scores []float64
	for f, err := range it {
		if err != nil {
			return nil, err
		}
		scores = AppendQualScores(scores[:0], f.FqQual())
		sum = GrowLen(sum, len(scores))
		count = GrowLen(count, len(scores))
		for i, score := range scores {
			sum[i] += score
			count[i]++
		}
	}
	for i, _ := range sum {
		sum[i] = sum[i] / count[i]
	}
	return sum, nil
}

func FullQualPerPos() {
	it := ParseFastq(os.Stdin)
	quals, err := QualPerPos(it)
	if err != nil {
		panic(err)
	}
	enc := json.NewEncoder(os.Stdout)
	err = enc.Encode(quals)
	if err != nil {
		panic(err)
	}
}
