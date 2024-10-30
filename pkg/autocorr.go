package fastats

import (
	"iter"

	"github.com/jgbaldwinbrown/iterh"
	"github.com/montanaflynn/stats"
)

// Requres sorted bed entries
func WindowBed[B BedEnter[T], T any](it iter.Seq[B], winsize, winstep int) iter.Seq[BedEntry[[]T]] {
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

// Requres sorted bed entries
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
