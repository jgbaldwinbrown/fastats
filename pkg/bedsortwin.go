package fastats

import (
	"iter"
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
)

func ShouldFinishWin(win ChrSpan, next ChrSpan) bool {
	return win.Chr != "" && (
		win.Chr != next.Chr || (
			win.End <= next.Start))
}

func FullyLeftOf(x, y ChrSpan) bool {
	return y.Start >= x.End || x.Chr != y.Chr
}

func toChrSpan[C ChrSpanner](c C) ChrSpan {
	return ChrSpan{Span: Span{Start: c.SpanStart(), End: c.SpanEnd()}, Chr: c.SpanChr()}
}

func CheckAndPopMulti[B BedEnter[T], T any](win ChrSpan, d *Deque[B]) {
	for d.Len() > 0 {
		next := d.Get(0)
		if FullyLeftOf(toChrSpan(next), win) {
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

func WindowSortedBed[B BedEnter[FT], FT any](it iter.Seq2[B, error], winsize, winstep int) func(func(BedEntry[[]B], error) bool) {
	return func(yield func(BedEntry[[]B], error) bool) {
		var d Deque[B]
		win := ChrSpan{"", Span{-1, -1}}
		out := BedEntry[[]B]{}
		// log.Printf("starting windowSortedBed")
		it(func(v B, e error) bool {
			if e != nil {
				yield(out, e)
				return false
			}
			cs := toChrSpan(v)
			for ShouldUpdateWin(win, cs) {
				// log.Printf("need to update win %v; v: %v\n", win, v)
				if ShouldFinishWin(win, cs) {
					log.Printf("should finish win; win %v; v %v\n", win, v)
					CheckAndPopMulti(win, &d)
					log.Printf("finished popping; d.Len() %v\n", d.Len())
					out.ChrSpan = win
					out.Fields = d.AppendToSlice(out.Fields[:0])
					// out.Fields = AppendDequeBedFields(out.Fields[:0], &d)
					// log.Printf("yielding %v\n", out)
					if ok := yield(out, nil); !ok {
						return false
					}
				}
				win = UpdateWin(win, v.SpanChr(), winsize, winstep)
				// log.Printf("updated win to %v\n", win)
			}
			d.PushBack(v)
			// log.Printf("pushed %v\n", v)
			return true
		})
		if win.Chr != "" {
			CheckAndPopMulti(win, &d)
			out.ChrSpan = win
			out.Fields = d.AppendToSlice(out.Fields[:0])
			// out.Fields = AppendDequeBedFields(out.Fields[:0], &d)
			// log.Printf("yielding %v\n", out)
			yield(out, nil)
		}
	}
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

func MeanBedPerBp[B BedEnter[float64]](bed []B) float64 {
	sum := 0.0
	count := 0.0
	for _, b := range bed {
		if !IsNaNOrInf(b.BedFields()) {
			sum += b.BedFields()
			count += float64(b.SpanEnd() - b.SpanStart())
		}
	}
	// fmt.Fprintf(os.Stderr, "got sum: %v count %v\n", sum, count)
	return sum / count
}

func MeanWindowCounts[B BedEnter[float64]](it iter.Seq2[B, error], winsize, winstep int) iter.Seq2[BedEntry[float64], error] {
	return func(yield func(BedEntry[float64], error) bool) {
		wins := WindowSortedBed(it, winsize, winstep)
		wins(func(win BedEntry[[]B], err error) bool {
			ok := yield(BedEntry[float64]{win.ChrSpan, MeanBedPerBp(win.Fields)}, err)
			return ok && err == nil
		})
	}
}

func ChrSpanLess[C ChrSpanner](x, y ChrSpanner) bool {
	xc := toChrSpan(x)
	yc := toChrSpan(y)
	if xc.Chr < yc.Chr {
		return true
	}
	if xc.Chr > yc.Chr {
		return false
	}
	if xc.Start < yc.Start {
		return true
	}
	if xc.Start > yc.Start {
		return false
	}
	if xc.End < yc.End {
		return true
	}
	return false
}

func SortBed[C ChrSpanner](bed []C) {
	sort.Slice(bed, func(i, j int) bool {
		return ChrSpanLess[C](bed[i], bed[j])
	})
}

func Collect[T any](it iter.Seq[T]) []T {
	var out []T
	it(func(val T) bool {
		out = append(out, val)
		return true
	})
	return out
}

func CollectErr[T any](it iter.Seq2[T, error]) ([]T, error) {
	var out []T
	var err error
	it(func(val T, e error) bool {
		if e != nil {
			err = e
			return false
		}
		out = append(out, val)
		return true
	})
	return out, err
}

func SortedBed[B ChrSpanner](it iter.Seq2[B, error]) ([]B, error) {
	sl, e := CollectErr[B](it)
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

func WriteFloatBedEntry[B BedEnter[float64]](w io.Writer, b B) (n int, err error) {
	return fmt.Fprintf(w, "%v\t%v\t%v\t%v\n", b.SpanChr(), b.SpanStart(), b.SpanEnd(), b.BedFields())
}

func WriteFloatBed[B BedEnter[float64]](w io.Writer, it iter.Seq2[B, error]) (n int, err error) {
	n = 0
	it(func(b B, e error) bool {
		if e != nil {
			err = e
			return false
		}
		nwritten, e := WriteFloatBedEntry(w, b)
		n += nwritten
		// log.Printf("wrote %v\n", b)
		if e != nil {
			err = e
			return false
		}
		return true
	})
	return n, err
}

func FlatToFloatBed[B BedEnter[[]string]](it iter.Seq2[B, error]) iter.Seq2[BedEntry[float64], error] {
	return func(yield func(BedEntry[float64], error) bool) {
		it(func(b B, err error) bool {
			if err != nil {
				var val BedEntry[float64]
				yield(val, err)
				return false
			}
			if len(b.BedFields()) < 1 {
				log.Printf("len(b.Fields) < 1; b %v\n", b)
				return yield(BedEntry[float64]{toChrSpan(b), math.NaN()}, nil)
			}
			x, e := strconv.ParseFloat(b.BedFields()[0], 64)
			if e != nil {
				log.Printf("b.Fields[0] not parsed; b %v\n", b)
				return yield(BedEntry[float64]{toChrSpan(b), math.NaN()}, nil)
			}
			return yield(BedEntry[float64]{toChrSpan(b), x}, nil)
		})
	}
}

func SliceIter2[S ~[]T, T any](s S) func(func(T, error) bool) {
	return func(yield func(T, error) bool) {
		for _, val := range s {
			if ok := yield(val, nil); !ok {
				return
			}
		}
	}
}

func BedSortWin(r io.Reader, w io.Writer, f BedSortWinFlags) error {
	// log.Print("started flat to floatbed")
	it := FlatToFloatBed(ParseBedFlat(r))
	// log.Print("finished flat to floatbed")

	if !f.Sorted {
		// log.Print("sorting")
		bed, e := SortedBed(it)
		if e != nil {
			return e
		}
		it = SliceIter2(bed)
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
