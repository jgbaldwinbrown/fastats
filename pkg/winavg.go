package fastats

import (
	"flag"
	"fmt"
	"github.com/montanaflynn/stats"
	"iter"
	"log"
	"os"
)

// func WindowSortedBed[B BedEnter[FT], FT any](it iter.Seq2[B, error], winsize, winstep int) func(func(BedEntry[[]B], error) bool) {

func WinAvgSortedBed[B BedEnter[float64]](it iter.Seq2[B, error], winsize, winstep int) iter.Seq2[BedEntry[float64], error] {
	return func(yield func(BedEntry[float64], error) bool) {
		winIter := WindowSortedBed(it, winsize, winstep)
		for win, err := range winIter {
			vals := []float64{}
			for _, bedEntry := range win.Fields {
				vals = append(vals, bedEntry.BedFields())
			}
			avg, err2 := stats.Mean(vals)
			if err2 != nil {
				log.Print(err2)
			}
			b := BedEntry[float64]{}
			b.ChrSpan = win.ChrSpan
			b.Fields = avg
			if !yield(b, err) {
				return
			}
		}
	}
}

type WinAvgSortedBedFlags struct {
	WinSize int
	WinStep int
}

func FullWinAvgSortedBed() {
	var f WinAvgSortedBedFlags
	flag.IntVar(&f.WinSize, "w", 1, "Set the window size")
	flag.IntVar(&f.WinStep, "s", 1, "Set the window step")
	flag.Parse()

	bed := ParseBedGraph(os.Stdin)
	wins := WinAvgSortedBed(bed, f.WinSize, f.WinStep)
	for win, err := range wins {
		if err != nil {
			log.Fatal(err)
		}
		_, err2 := fmt.Printf("%v\t%v\t%v\t%v\n", win.Chr, win.Start, win.End, win.Fields)
		if err2 != nil {
			log.Fatal(err2)
		}
	}
}

func WinSumSortedBed[B BedEnter[float64]](it iter.Seq2[B, error], winsize, winstep int, perBp bool) iter.Seq2[BedEntry[float64], error] {
	return func(yield func(BedEntry[float64], error) bool) {
		winIter := WindowSortedBed(it, winsize, winstep)
		for win, err := range winIter {
			vals := []float64{}
			for _, bedEntry := range win.Fields {
				vals = append(vals, bedEntry.BedFields())
			}
			avg, err2 := stats.Sum(vals)
			if err2 != nil {
				log.Print(err2)
			}
			if perBp {
				avg /= float64(win.SpanEnd() - win.SpanStart())
			}
			b := BedEntry[float64]{}
			b.ChrSpan = win.ChrSpan
			b.Fields = avg
			if !yield(b, err) {
				return
			}
		}
	}
}

type WinSumSortedBedFlags struct {
	WinSize int
	WinStep int
	PerBp bool
}

func FullWinSumSortedBed() {
	var f WinSumSortedBedFlags
	flag.IntVar(&f.WinSize, "w", 1, "Set the window size")
	flag.IntVar(&f.WinStep, "s", 1, "Set the window step")
	flag.BoolVar(&f.PerBp, "b", false, "Divide sum by size of window")
	flag.Parse()

	bed := ParseBedGraph(os.Stdin)
	wins := WinSumSortedBed(bed, f.WinSize, f.WinStep, f.PerBp)
	for win, err := range wins {
		if err != nil {
			log.Fatal(err)
		}
		_, err2 := fmt.Printf("%v\t%v\t%v\t%v\n", win.Chr, win.Start, win.End, win.Fields)
		if err2 != nil {
			log.Fatal(err2)
		}
	}
}
