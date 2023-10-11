package fastats

import (
	"fmt"
	"os"
	"flag"
	"bufio"
)

func AddKmers(kmap map[string]int64, k int, seq string) {
	for i := 0; i + k <= len(seq); i++ {
		kmap[seq[i:i+k]]++
	}
}

func CountKmers(it Iter[FaEntry], k int) (map[string]int64, error) {
	m := map[string]int64{}
	err := it.Iterate(func(f FaEntry) error {
		AddKmers(m, k, f.Seq)
		return nil
	})
	return m, err
}

type Kmer struct {
	Seq string
	Count int64
}

func KmerIter(m map[string]int64) *Iterator[Kmer] {
	return &Iterator[Kmer]{Iteratef: func(yield func(Kmer) error) error {
		for seq, count := range m {
			if e := yield(Kmer{seq, count}); e != nil {
				return e
			}
		}
		return nil
	}}
}

func KmerHist(it Iter[Kmer]) ([]int64, error) {
	var out []int64
	e := it.Iterate(func(k Kmer) error {
		out = GrowLen(out, int(k.Count) + 1)
		out[k.Count]++
		return nil
	})
	if e != nil {
		return nil, e
	}
	return out, nil
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

	hist, e := KmerHist(KmerIter(m))
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
