package fastats

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"iter"
	"os"
	"regexp"
	"sort"

	"github.com/jgbaldwinbrown/zfile"
)

// ID=SNP_1;Name=gap;subst_len=1;query_dir=1;query_sequence=3R;query_coord=12-12;query_bases=N;ref_bases=t;color=#42C042

type NucdiffAttr struct {
	ID         string
	Name       string
	Len        int
	QueryDir   int
	QuerySeq   string
	QueryStart int
	QueryEnd   int
	QueryBases string
	RefBases   string
	Color      string
}

var nucdiffAttrRe = regexp.MustCompile(
	`^ID=([^;]*);Name=([^;]*);[^_]*_len=([^;]*);query_dir=([^;]*);query_sequence=([^;]*);query_coord=([^\-]*)-([^;]*);query_bases=([^;]*);color=(.*)$`,
)

func ParseNucdiffAttr(in string) (NucdiffAttr, error) {
	fields := nucdiffAttrRe.FindStringSubmatch(in)
	var n NucdiffAttr
	if len(fields) < 11 {
		return n, fmt.Errorf("len(fields) %v < 11", len(fields))
	}
	var e error
	f := fields

	n.ID = f[1]
	n.Name = f[2]
	n.QuerySeq = f[5]
	n.QueryBases = f[8]
	n.RefBases = f[9]
	n.Color = f[10]
	if n.Len, e = strconv.Atoi(f[3]); e != nil {
		return n, e
	}
	if n.QueryDir, e = strconv.Atoi(f[4]); e != nil {
		return n, e
	}
	if n.QueryStart, e = strconv.Atoi(f[6]); e != nil {
		return n, e
	}
	if n.QueryEnd, e = strconv.Atoi(f[7]); e != nil {
		return n, e
	}
	return n, nil
}

type NucdiffGT struct {
	Ref    int
	Alt    int
	Phased bool
}

type NucdiffData struct {
	CrossNames []string
	M          map[ChrSpan]map[string]NucdiffAttr
}

func NucdiffReadGff[G GffEnter[NucdiffAttr]](d *NucdiffData, crossname string, it iter.Seq2[G, error]) error {
	d.CrossNames = append(d.CrossNames, crossname)
	for g, e := range it {
		if e != nil {
			return e
		}

		m, ok := d.M[toChrSpan(g)]
		if !ok {
			m = map[string]NucdiffAttr{}
			d.M[toChrSpan(g)] = m
		}
		m[crossname] = g.GffAttributes()
	}
	return nil
}

func NucdiffReadGffs(paths []string) (*NucdiffData, error) {
	d := new(NucdiffData)
	d.M = map[ChrSpan]map[string]NucdiffAttr{}

	for _, path := range paths {
		err := func() error {
			r, e := zfile.Open(path)
			if e != nil {
				return e
			}
			defer func() { Must(r.Close()) }()

			return NucdiffReadGff(d, path, ParseGff(r, ParseNucdiffAttr))
		}()
		if err != nil {
			return nil, err
		}
	}

	return d, nil
}

type StringFormatter string

func (f StringFormatter) Format(format string) string {
	return string(f)
}

// func MakeNucdiffVCFEntry(d *NucdiffData, c ChrSpan) (VcfEntry[StructuredInfoSamples[InfoPair[string], StringFormatter]], error) {
// 	v := VcfEntry[StructuredInfoSamples[InfoPair[string], StringFormatter]]{}
// 	v.ChrSpan = c
//
// }

func NucdiffReadVcf[V VcfHeader](d *NucdiffVcfData[V], crossname string, it iter.Seq2[V, error]) error {
	d.CrossNames = append(d.CrossNames, crossname)
	for v, e := range it {
		if e != nil {
			return e
		}
		m, ok := d.M[toChrSpan(v)]
		if !ok {
			m = map[string]V{}
			d.M[toChrSpan(v)] = m
		}
		m[crossname] = v
	}
	return nil
}

func NucdiffReadVcfs(refname string, crossnames []string, paths []string) (*NucdiffVcfData[VcfEntry[struct{}]], error) {
	if len(crossnames) != len(paths) {
		return nil, fmt.Errorf("len(crossnames) %v != len(paths) %v", len(crossnames), len(paths))
	}

	d := new(NucdiffVcfData[VcfEntry[struct{}]])
	d.RefName = refname
	d.M = map[ChrSpan]map[string]VcfEntry[struct{}]{}

	for i, path := range paths {
		err := func() error {
			r, e := zfile.Open(path)
			if e != nil {
				return e
			}
			defer func() { Must(r.Close()) }()

			return NucdiffReadVcf[VcfEntry[struct{}]](d, crossnames[i], ParseSimpleVcf(r))
		}()
		if err != nil {
			return nil, err
		}
	}

	return d, nil
}

type NucdiffVcfData[V VcfHeader] struct {
	RefName    string
	CrossNames []string
	M          map[ChrSpan]map[string]V
}

func GetSortedChrSpans[V VcfHeader](d *NucdiffVcfData[V]) []ChrSpan {
	cspans := make([]ChrSpan, 0, len(d.M))
	for k, _ := range d.M {
		cspans = append(cspans, k)
	}
	sort.Slice(cspans, func(i, j int) bool { return ChrSpanLess[ChrSpan](cspans[i], cspans[j]) })
	return cspans
}

func GetAlleles[V VcfHeader](crossnames []string, dmap map[string]V) (alleles []string, crossToIndex map[string]int) {
	m := map[string]int{}
	alleleset := map[string]int{}
	for crossname, dv := range dmap {
		a := dv.VcfAlts()[0]
		idx, ok := alleleset[a]
		if !ok {
			alleles = append(alleles, a)
			idx = len(alleles)
			alleleset[a] = idx
		}
		m[crossname] = idx
	}
	return alleles, m
}

func MergeNucdiffVcfEntries[V VcfHeader](d *NucdiffVcfData[V], refname string, crossnames []string, crossidxs map[string]int, chrspan ChrSpan) VcfEntry[StructuredInfoSamples[string, StringFormatter]] {
	var v VcfEntry[StructuredInfoSamples[string, StringFormatter]]
	dmap, ok := d.M[chrspan]
	if !ok {
		panic(fmt.Errorf("d.M does not contain chrspan %v", chrspan))
	}

	v.InfoAndSamples.Info = nil
	v.InfoAndSamples.Samples.Format = []string{"GT"}

	for i := -1; i < len(crossnames); i++ {
		v.InfoAndSamples.Samples.Samples = append(v.InfoAndSamples.Samples.Samples, []StringFormatter{StringFormatter("0/0")})
	}

	var allelemap map[string]int
	v.Alts, allelemap = GetAlleles(crossnames, dmap)

	for crossname, dv := range dmap {
		v.ChrSpan = toChrSpan(dv)
		v.ID = dv.VcfID()
		v.Ref = dv.VcfRef()
		v.Qual = dv.VcfQual()
		v.Filter = dv.VcfFilter()
		altidx, ok := allelemap[crossname]
		if ok {
			v.InfoAndSamples.Samples.Samples[crossidxs[crossname]] = []StringFormatter{StringFormatter(fmt.Sprintf("%v/%v", altidx, altidx))}
		}
	}

	return v
}

func NucdiffWriteVcf[V VcfHeader](d *NucdiffVcfData[V], w io.Writer) error {
	cw := csv.NewWriter(w)
	cw.Comma = rune('\t')
	defer func() { cw.Flush() }()

	chrspans := GetSortedChrSpans(d)

	crossidxs := map[string]int{}
	crossidxs[d.RefName] = 0
	for i, name := range d.CrossNames {
		crossidxs[name] = i + 1
	}

	_, e := fmt.Fprintf(w, "#CHROM  POS     ID      REF     ALT     QUAL    FILTER	%v", d.RefName)
	if e != nil {
		return e
	}

	for _, crossname := range d.CrossNames {
		_, e = fmt.Fprintf(w, "\t%v", crossname)
		if e != nil {
			return e
		}
	}

	_, e = fmt.Fprintf(w, "\n")
	if e != nil {
		return e
	}

	var c []string
	for _, cs := range chrspans {
		v := MergeNucdiffVcfEntries(d, d.RefName, d.CrossNames, crossidxs, cs)

		c, e := StructuredVcfEntryToCsv(c[:0], v)
		if e != nil {
			return e
		}

		e = cw.Write(c)
		if e != nil {
			return e
		}
	}

	return nil
}

type NucdiffVcfPair struct {
	Name string
	Path string
}

func ReadNucdiffVcfPairs(r io.Reader) ([]NucdiffVcfPair, error) {
	var out []NucdiffVcfPair
	var p NucdiffVcfPair
	dec := json.NewDecoder(r)

	for {
		e := dec.Decode(&p)
		if e == io.EOF {
			break
		}
		if e != nil {
			return nil, e
		}
		out = append(out, p)
	}
	return out, nil
}

func NucdiffFullVcfMerge() {
	rname := flag.String("r", "", "Reference name (required)")
	flag.Parse()
	if *rname == "" {
		panic(fmt.Errorf("missing -r"))
	}

	pairs, e := ReadNucdiffVcfPairs(os.Stdin)
	Must(e)

	crossnames := make([]string, 0, len(pairs))
	paths := make([]string, 0, len(pairs))
	for _, pair := range pairs {
		crossnames = append(crossnames, pair.Name)
		paths = append(paths, pair.Path)
	}

	d, e := NucdiffReadVcfs(*rname, crossnames, paths)
	Must(e)

	Must(NucdiffWriteVcf(d, os.Stdout))
}
