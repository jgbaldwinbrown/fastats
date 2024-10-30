package fastats

import (
	"iter"

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
