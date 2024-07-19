package fastats

import (
	"fmt"
	"os"
	"flag"
	"bufio"
	"iter"
)

func AddKmers(kmap map[string]int64, k int, seq string) {
	for i := 0; i + k <= len(seq); i++ {
		kmap[seq[i:i+k]]++
	}
}

func CountKmers[F FaEnter](it iter.Seq2[F, error], k int) (map[string]int64, error) {
	m := map[string]int64{}
	for f, err := range it {
		if err != nil {
			return m, err
		}
		AddKmers(m, k, f.FaSeq())
	}
	return m, nil
}

type Kmer struct {
	Seq string
	Count int64
}

func KmerIter(m map[string]int64) iter.Seq[Kmer] {
	return func(yield func(Kmer) bool) {
		for seq, count := range m {
			if !yield(Kmer{seq, count}) {
				return
			}
		}
	}
}

func KmerHist(it iter.Seq[Kmer]) []int64 {
	var out []int64
	for k := range it {
		out = GrowLen(out, int(k.Count) + 1)
		out[k.Count]++
	}
	return out
}

func FullCountKmers() {
	k := flag.Int("k", 1, "k")
	flag.Parse()

	it := ParseFasta(os.Stdin)
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	m, e := CountKmers(it, *k)
	if e != nil {
		panic(e)
	}

	for key, val := range m {
		_, e := fmt.Fprintf(w, "%v\t%v\n", key, val)
		if e != nil {
			panic(e)
		}
	}
}

func FullKmerHist() {
	k := flag.Int("k", 1, "k")
	flag.Parse()

	it := ParseFasta(os.Stdin)
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	m, e := CountKmers(it, *k)
	if e != nil {
		panic(e)
	}

	hist := KmerHist(KmerIter(m))
	if e != nil {
		panic(e)
	}

	for count, freq := range hist {
		_, e := fmt.Fprintf(w, "%v\t%v\n", count, freq)
		if e != nil {
			panic(e)
		}
	}
}
