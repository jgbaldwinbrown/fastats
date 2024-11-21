package fastats

import (
	"iter"
	"flag"
	"os"
	"log"
	"fmt"
	"strconv"

	"github.com/jgbaldwinbrown/iterh"
	"github.com/montanaflynn/stats"
)

// Requires sorted bed entries
func SplitChrs[C ChrSpanner](cit iter.Seq[C]) iter.Seq[iter.Seq[C]] {
	return func(y1 func(iter.Seq[C]) bool) {
		it := iterh.AddNilError(cit)
		m, _ := CollectChrSpannerMap(it)
		for _, spans := range m {
			f := func(y2 func(C) bool) {
				for _, span := range spans {
					if !y2(span) {
						return
					}
				}
			}
			if !y1(f) {
				return
			}
		}
	}
}

// // Requires sorted bed entries
// func SplitChrs[C Chrer](it iter.Seq[C]) iter.Seq[iter.Seq[C]] {
// 	return func(y1 func(iter.Seq[C] bool)) {
// 		p, cancel := iter.Pull(it)
// 		defer cancel()
// 		for val, ok := p(); ok; val, ok = p() {
// 			f := func(y2 func(C) bool) {
// 				if !y2(p) {
// 					return
// 				}
// 				for val, ok = p(); ok; val, ok = p() {
// 					
// 				}
// 			}
// 		}
// 	}
// }

// Requires sorted bed entries
func WindowBed[B BedEnter[T], T any](it iter.Seq[B], winsize, winstep int) iter.Seq[BedEntry[[]T]] {
	return func(y func(BedEntry[[]T]) bool) {
		chrBeds := SplitChrs(it)
		for chrBed := range chrBeds {
			windowed := WindowBedWeak(chrBed, winsize, winstep)
			for b := range windowed {
				if !y(b) {
					return
				}
			}
		}
	}
}

// Requires sorted bed entries
func WindowBedWeak[B BedEnter[T], T any](it iter.Seq[B], winsize, winstep int) iter.Seq[BedEntry[[]T]] {
	return func(y func(BedEntry[[]T]) bool) {
		s := SpreadBed(it)
		wins := iterh.Window(s, winsize, winstep)
		for win := range wins {
			length := win.Len()
			if length < 1 {
				continue
			}
			outb := BedEntry[[]T]{}
			outb.ChrSpan = ToChrSpan(win.At(0))
			outb.End = win.At(length - 1).SpanEnd()
			outb.Fields = make([]T, 0, length)
			for i := 0; i < length; i++ {
				outb.Fields = append(outb.Fields, win.At(i).BedFields())
			}
			if !y(outb) {
				return
			}
		}
	}
}

// Requires sorted bed entries
func AutoCorrelationWindows[B BedEnter[float64]](it iter.Seq[B], lags, winsize, winstep int) iter.Seq[BedEntry[float64]] {
	return func(y func(BedEntry[float64]) bool) {
		wb := WindowBed(it, winsize, winstep)
		for win := range wb {
			if len(win.Fields) < 1 {
				continue
			}
			corr, _ := stats.AutoCorrelation(win.Fields, lags)
			out := BedEntry[float64]{}
			out.ChrSpan = ToChrSpan(win)
			out.Fields = corr
			if !y(out) {
				return
			}
		}
	}
}

type AutoCorrelationFlags struct {
	Lag int
	Winsize int
	Winstep int
	Col int
}

func ColToFloat(col int) func([]string) (float64, error) {
	return func(fields []string) (float64, error) {
		if len(fields) <= col {
			return 0, fmt.Errorf("ColToFloat: col %v < len(fields); fields %v", col, fields)
		}
		val, e := strconv.ParseFloat(fields[col], 64)
		return val, e
	}
}

func ToBedGraphEntry[B BedEnter[[]string]](col int) func(B) (BedEntry[float64], error) {
	return func(b B) (BedEntry[float64], error) {
		var ent BedEntry[float64]
		ent.ChrSpan = ToChrSpan(b)
		fields := b.BedFields()
		if len(fields) <= col {
			return ent, fmt.Errorf("ToBedGraphEntry: col %v < len(fields); fields %v", col, fields)
		}
		var e error
		ent.Fields, e = strconv.ParseFloat(fields[col], 64)
		return ent, e
	}
}

func FullAutoCorrelationWindows() {
	var f AutoCorrelationFlags
	flag.IntVar(&f.Lag, "l", 1, "Lag")
	flag.IntVar(&f.Winsize, "w", 10, "Window size")
	flag.IntVar(&f.Winstep, "s", 1, "Window step")
	flag.IntVar(&f.Col, "c", 0, "Field column to correlate")
	flag.Parse()

	bed, errp := iterh.BreakWithError(ParseBed(os.Stdin, ColToFloat(f.Col)))
	
	a := AutoCorrelationWindows(bed, f.Lag, f.Winsize, f.Winstep)
	for win := range a {
		_, e := fmt.Printf("%v\t%v\t%v\t%v\n", win.Chr, win.Start, win.End, win.Fields)
		if e != nil {
			log.Fatal(e)
		}
	}
	if *errp != nil {
		log.Fatal(*errp)
	}
}
