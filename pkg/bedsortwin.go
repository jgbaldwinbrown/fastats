package fastats

import (
	"log"
	"fmt"
	"strconv"
	"flag"
	"sort"
	"io"
	"bufio"
	"os"
	"math"
	"github.com/montanaflynn/stats"
	"github.com/jgbaldwinbrown/iter"
)

func ShouldFinishWin(win ChrSpan, next ChrSpan) bool {
	return win.Chr != "" && (
		win.Chr != next.Chr || (
			win.End <= next.Start))
}

func FullyLeftOf(x, y ChrSpan) bool {
	return y.Start >= x.End || x.Chr != y.Chr
}

func CheckAndPopMulti[T any](win ChrSpan, d *Deque[BedEntry[T]]) {
	for d.Len() > 0 {
		next := d.Get(0)
		if FullyLeftOf(next.ChrSpan, win) {
			// log.Printf("popping %v; win %v\n", next, win)
			d.PopFront()
		} else {
			return
		}
	}
}

func ShouldUpdateWin(win, next ChrSpan) bool {
	return win.Chr != next.Chr || win.End <= next.Start
}

func AppendDequeBedFields[FT any](s []FT, d *Deque[BedEntry[FT]]) []FT {
	l := d.Len()
	for i := 0; i < l; i++ {
		val := d.Get(i)
		s = append(s, val.Fields)
	}
	return s
}

func UpdateWin(win ChrSpan, chr string, winsize, winstep int) ChrSpan {
	if win.Chr != chr {
		win.Chr = chr
		win.Start = 0
	} else {
		win.Start += int64(winstep)
	}
	win.End = win.Start + int64(winsize)
	return win
}

func WindowSortedBed[FT any](it iter.Iter[BedEntry[FT]], winsize, winstep int) *iter.Iterator[BedEntry[[]BedEntry[FT]]] {
	return &iter.Iterator[BedEntry[[]BedEntry[FT]]]{Iteratef: func(yield func(BedEntry[[]BedEntry[FT]]) error) error {
		var d Deque[BedEntry[FT]]
		win := ChrSpan{"", Span{-1, -1}}
		out := BedEntry[[]BedEntry[FT]]{}
		// log.Printf("starting windowSortedBed")
		it.Iterate(func(v BedEntry[FT]) error {
			for ShouldUpdateWin(win, v.ChrSpan) {
				// log.Printf("need to update win %v; v: %v\n", win, v)
				if ShouldFinishWin(win, v.ChrSpan) {
					log.Printf("should finish win; win %v; v %v\n", win, v)
					CheckAndPopMulti(win, &d)
					log.Printf("finished popping; d.Len() %v\n", d.Len())
					out.ChrSpan = win
					out.Fields = d.AppendToSlice(out.Fields[:0])
					// out.Fields = AppendDequeBedFields(out.Fields[:0], &d)
					// log.Printf("yielding %v\n", out)
					if e := yield(out); e != nil {
						return e
					}
				}
				win = UpdateWin(win, v.Chr, winsize, winstep)
				// log.Printf("updated win to %v\n", win)
			}
			d.PushBack(v)
			// log.Printf("pushed %v\n", v)
			return nil
		})
		if win.Chr != "" {
			CheckAndPopMulti(win, &d)
			out.ChrSpan = win
			out.Fields = d.AppendToSlice(out.Fields[:0])
			// out.Fields = AppendDequeBedFields(out.Fields[:0], &d)
			// log.Printf("yielding %v\n", out)
			if e := yield(out); e != nil {
				return e
			}
		}
		return nil
	}}
}

func NoNaNs(x []float64) []float64 {
	out := make([]float64, 0, len(x))
	for _, v := range x {
		if !math.IsNaN(v) && !math.IsInf(v, 0) {
			out = append(out, v)
		}
	}
	return out
}

func MeanNoNaN(x []float64) float64 {
	nonan := NoNaNs(x)
	m, e := stats.Mean(nonan)
	if e != nil {
		return math.NaN()
	}
	return m
}

func IsNaNOrInf(x float64) bool {
	return math.IsNaN(x) || math.IsInf(x, 0)
}

func MeanBedPerBp(bed []BedEntry[float64]) float64 {
	sum := 0.0
	count := 0.0
	for _, b := range bed {
		if !IsNaNOrInf(b.Fields) {
			sum += b.Fields
			count += float64(b.End - b.Start)
		}
	}
	// fmt.Fprintf(os.Stderr, "got sum: %v count %v\n", sum, count)
	return sum / count
}

func MeanWindowCounts(it iter.Iter[BedEntry[float64]], winsize, winstep int) (*iter.Iterator[BedEntry[float64]]) {
	return &iter.Iterator[BedEntry[float64]]{Iteratef: func(yield func(BedEntry[float64]) error) error {
		wins := WindowSortedBed[float64](it, winsize, winstep)
		return wins.Iterate(func(win BedEntry[[]BedEntry[float64]]) error {
			// if len(win.Fields) < 1 {
			// 	log.Printf("len(win.Fields) < 1; win %v\n", win)
			// }
			return yield(BedEntry[float64]{win.ChrSpan, MeanBedPerBp(win.Fields)})
		})
	}}
}

func ChrSpanLess(x, y ChrSpan) bool {
	if x.Chr < y.Chr {
		return true
	}
	if x.Chr > y.Chr {
		return false
	}
	if x.Start < y.Start {
		return true
	}
	if x.Start > y.Start {
		return false
	}
	if x.End < y.End {
		return true
	}
	return false
}

func SortBed[T any](bed []BedEntry[T]) {
	sort.Slice(bed, func(i, j int) bool {
		return ChrSpanLess(bed[i].ChrSpan, bed[j].ChrSpan)
	})
}

func SortedBed[T any](it iter.Iter[BedEntry[T]]) ([]BedEntry[T], error) {
	sl, e := iter.Collect[BedEntry[T]](it)
	if e != nil {
		return nil, e
	}
	SortBed(sl)
	return sl, nil
}

type BedSortWinFlags struct {
	Sorted bool
	Winsize int
	Winstep int
}

func WriteFloatBedEntry(w io.Writer, b BedEntry[float64]) (n int, err error) {
	return fmt.Fprintf(w, "%v\t%v\t%v\t%v\n", b.Chr, b.Start, b.End, b.Fields)
}

func WriteFloatBed(w io.Writer, it iter.Iter[BedEntry[float64]]) (n int, err error) {
	n = 0
	err = it.Iterate(func(b BedEntry[float64]) error {
		nwritten, e := WriteFloatBedEntry(w, b)
		n += nwritten
		// log.Printf("wrote %v\n", b)
		return e
	})
	return n, err
}

func FlatToFloatBed(it iter.Iter[BedEntry[[]string]]) *iter.Iterator[BedEntry[float64]] {
	return &iter.Iterator[BedEntry[float64]]{Iteratef: func(yield func(BedEntry[float64]) error) error {
		return it.Iterate(func(b BedEntry[[]string]) error {
			if len(b.Fields) < 1 {
				log.Printf("len(b.Fields) < 1; b %v\n", b)
				return yield(BedEntry[float64]{b.ChrSpan, math.NaN()})
			}
			x, e := strconv.ParseFloat(b.Fields[0], 64)
			if e != nil {
				log.Printf("b.Fields[0] not parsed; b %v\n", b)
				return yield(BedEntry[float64]{b.ChrSpan, math.NaN()})
			}
			return yield(BedEntry[float64]{b.ChrSpan, x})
		})
	}}
}

func BedSortWin(r io.Reader, w io.Writer, f BedSortWinFlags) error {
	// log.Print("started flat to floatbed")
	it := FlatToFloatBed(ParseBedFlat(r))
	// log.Print("finished flat to floatbed")

	if !f.Sorted {
		// log.Print("sorting")
		bed, e := SortedBed[float64](it)
		if e != nil {
			return e
		}
		it = iter.SliceIter[BedEntry[float64]](bed)
		// log.Print("sorted")
	}

	// log.Print("starting meanwindowcounts")
	wins := MeanWindowCounts(it, f.Winsize, f.Winstep)
	if _, e := WriteFloatBed(w, wins); e != nil {
		return e
	}
	return nil
}

func RunBedSortWin() {
	var f BedSortWinFlags
	flag.BoolVar(&f.Sorted, "sorted", false, "bed input already sorted")
	flag.IntVar(&f.Winsize, "size", 1, "Window size")
	flag.IntVar(&f.Winstep, "step", 1, "Window step")
	flag.Parse()

	stdin := bufio.NewReader(os.Stdin)
	stdout := bufio.NewWriter(os.Stdout)
	defer func() {
		e := stdout.Flush()
		if e != nil {
			log.Fatal(e)
		}
	}()

	if e := BedSortWin(stdin, stdout, f); e != nil {
		log.Fatal(e)
	}
}
