package fastats

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"iter"
	"log"
	"os"
	"regexp"

	"github.com/jgbaldwinbrown/iterh"
)

type gcRegionFlags struct {
	RegionsPath string
	FastaPath string
}

var gzRe = regexp.MustCompile(`\.gz$`)
var bedRe = regexp.MustCompile(`\.bed$`)
var gffRe = regexp.MustCompile(`\.gff$`)
var gtfRe = regexp.MustCompile(`\.gtf$`)

func toChrSpannerIter[C ChrSpanner](f func(io.Reader) iter.Seq2[C, error]) func(io.Reader) iter.Seq2[ChrSpanner, error] {
	return func(r io.Reader) iter.Seq2[ChrSpanner, error] {
		return func(yield func(ChrSpanner, error) bool) {
			for c, e := range f(r) {
				if !yield(c, e) {
					return
				}
			}
		}
	}
}

func ParseChrSpanner(path string) iter.Seq2[ChrSpanner, error] {
	opener := iterh.PathIter[ChrSpanner]
	if gzRe.MatchString(path) {
		opener = iterh.GzPathIter
	}
	var parser func(io.Reader) iter.Seq2[ChrSpanner, error]
	stripped := gzRe.ReplaceAllString(path, "")
	if bedRe.MatchString(stripped) {
		parser = toChrSpannerIter(ParseBedFlat)
	} else if gffRe.MatchString(stripped) || gtfRe.MatchString(stripped) {
		parser = toChrSpannerIter(ParseGffFlat)
	}
	if parser == nil {
		log.Fatal("ParseChrSpanner: path does not match .bed, .gff, or .gtf")
	}
	return opener(path, parser)
}

type ChrStat struct {
	Chr string
	Stat float64
}

func GcChrSpanFa[C ChrSpanner, F FaEnter](it iter.Seq2[ChrSpanFa[C, F], error]) iter.Seq2[ChrStat, error] {
	return func(yield func(ChrStat, error) bool) {
		for csf, e := range it {
			if e != nil {
				if !yield(ChrStat{}, e) {
					return
				}
				log.Print("GcChrSpanFa non-fatal error: %w", e)
				continue
			}
			out, e := ExtractOne(ToFaEntry(csf.FaEnter), ToSpan(csf.ChrSpanner))
			if e != nil {
				if !yield(ChrStat{}, e) {
					return
				}
				log.Print("GcChrSpanFa non-fatal error: %w", e)
				continue
			}
			if g, ok := any(csf.ChrSpanner).(GffEnter[[]AttributePair]); ok {
				for _, pair := range g.GffAttributes() {
					if pair.Tag == "ID" {
						out.Header = pair.Value + "\t" + out.Header
						break
					}
				}
			}
			if !yield(ChrStat{Chr: out.Header, Stat: GCFrac(out.Seq)}, nil) {
				return
			}
		}
	}
}

func FullGCRegions() {
	var f gcRegionFlags
	flag.StringVar(&f.RegionsPath, "r", "", "Path to (possibly gzipped) .bed, .gff or .gtf file containing regions")
	flag.StringVar(&f.FastaPath, "f", "", "Path to gzipped fasta file")
	flag.Parse()
	if f.RegionsPath == "" {
		log.Fatal("missing -r")
	}
	if f.FastaPath == "" {
		log.Fatal("missing -f")
	}

	fa := iterh.GzPathIter(f.FastaPath, ParseFasta)
	spans := ParseChrSpanner(f.RegionsPath)
	extracted := ExtractChrSpanFa(fa, spans)

	bw := bufio.NewWriter(os.Stdout)
	defer func() {
		e := bw.Flush()
		if e != nil {
			log.Fatal(e)
		}
	}()
	for cs, e := range GcChrSpanFa(extracted) {
		if e != nil {
			log.Fatal(e)
		}
		if _, e := fmt.Fprintf(bw, "%v\t%v\n", cs.Chr, cs.Stat); e != nil {
			log.Fatal(e)
		}
	}
}
