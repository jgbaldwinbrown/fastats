package fastats

import (
	"fmt"
	"strings"
	"strconv"

	"github.com/jgbaldwinbrown/csvh"
)

func ParseSamHeading(s string) SamEntry {
	return SamEntry{IsHeader: true, Header: s[1:]}
}

type SamAlignment struct {
	Qname string
	Flag uint16
	Rname string
	Pos int64
	Mapq int64
	CIGAR string
	Rnext string
	Pnext int64
	Tlen int64
	Seq string
	Qual string
}

type SamOptional struct {
	Tag [2]byte
	Type byte
	Char byte
	Int int64
	Float float64
	String string
	ByteArray []byte
	NumArrayType byte
	IntArray []int64
	FloatArray []float64
}

type SamEntry struct {
	IsHeader bool
	Header string
	SamAlignment
	Optional []SamOptional
}

func ParseSamAlignment(line []string) (SamAlignment, error) {
	var a SamAlignment
	_, e := csvh.Scan(line, &a.Qname, &a.Flag, &a.Rname, &a.Pos, &a.Mapq, &a.CIGAR, &a.Rnext, &a.Pnext, &a.Tlen, &a.Seq, &a.Qual)
	return a, e
}

func ParseByteArray(s string) ([]byte, error) {
	if len(s) % 2 != 0 {
		return nil, fmt.Errorf("ParseByteArray: len(s) % 2 != 0; len(s) %v; s %v", len(s), s)
	}
	out := make([]byte, 0, len(s) / 2)
	for i := 0; i < len(s); i += 2 {
		val := s[i:i+2]
		ival, e := strconv.ParseUint(val, 16, 64)
		if e != nil {
			return nil, e
		}
		out = append(out, byte(ival))
	}
	return out, nil
}

func ParseIntArray(fields []string) ([]int64, error) {
	out := make([]int64, 0, len(fields))
	for _, field := range fields {
		val, e := strconv.ParseInt(field, 0, 64)
		if e != nil {
			return nil, e
		}
		out = append(out, val)
	}
	return out, nil
}

func ParseFloatArray(fields []string) ([]float64, error) {
	out := make([]float64, 0, len(fields))
	for _, field := range fields {
		val, e := strconv.ParseFloat(field, 64)
		if e != nil {
			return nil, e
		}
		out = append(out, val)
	}
	return out, nil
}

func ParseNumArray(s string) (ntype byte, array any, err error) {
	fields := strings.Split(s, ",")
	if len(fields) < 1 {
		return 0, nil, fmt.Errorf("ParseNumArray: len(fields) %v < 1; fields %v", len(fields), fields)
	}
	if len(fields[0]) != 1 {
		return 0, nil, fmt.Errorf("ParseNumArray: len(fields[0]) %v != 1; fields[0] %v", len(fields[0]), fields[0])
	}
	ntype = fields[0][0]
	if ntype == 'f' {
		array, err = ParseFloatArray(fields[1:])
	} else {
		array, err = ParseIntArray(fields[1:])
	}
	return ntype, array, err
}

func ParseSamOptional(s string) (SamOptional, error) {
	var o SamOptional
	fields := strings.Split(s, ":")
	if len(fields) != 3 {
		return o, fmt.Errorf("ParseSamOptional: len(fields) %v != 3; fields %v", len(fields), fields)
	}
	tag, ttype, val := fields[0], fields[1], fields[2]
	if len(tag) != 2 {
		return o, fmt.Errorf("ParseSamOptional: len(tag) %v != 2; tag %v", len(tag), tag)
	}
	copy(o.Tag[:], tag)
	if len(ttype) != 1 {
		return o, fmt.Errorf("ParseSamOptional: len(type) %v != 1; type %v", len(ttype), ttype)
	}
	o.Type = ttype[0]
	var e error
	switch ttype {
	case "A":
		if len(val) != 1 {
			return o, fmt.Errorf("ParseSamOptional: type %v: len(val) %v != 1; val %v", ttype, len(val), val)
		}
		o.Char = val[0]
	case "i":
		o.Int, e = strconv.ParseInt(val, 0, 64)
	case "f":
		o.Float, e = strconv.ParseFloat(val, 64)
	case "Z":
		o.String = val
	case "H":
		o.ByteArray, e = ParseByteArray(val)
	case "B":
		var numArray any
		o.NumArrayType, numArray, e = ParseNumArray(val)
		if o.NumArrayType == 'f' {
			o.FloatArray = numArray.([]float64)
		} else {
			o.IntArray = numArray.([]int64)
		}
	default:
		return o, fmt.Errorf("ParseSamOptional: invalid type %v", ttype)
	}
	return o, e
}

func ParseSamOptionals(s string) ([]SamOptional, error) {
	fields := strings.Split(s, ",")
	out := make([]SamOptional, 0, len(fields))
	for _, field := range fields {
		opt, e := ParseSamOptional(field)
		if e != nil {
			return nil, e
		}
		out = append(out, opt)
	}
	return out, nil
}

func ParseSamEntry(s string) (SamEntry, error) {
	if len(s) < 0 && s[0] == '@' {
		return ParseSamHeading(s), nil
	}
	var a SamEntry
	var e error
	line := strings.Split(s, "\t")
	a.SamAlignment, e = ParseSamAlignment(line)
	if e != nil {
		return a, e
	}
	if len(line) < 12 {
		return a, nil
	}
	a.Optional, e = ParseSamOptionals(line[11])
	return a, e
}

// Col	Field	Type	Brief description
// 1	QNAME	String	Query template NAME
// 2	FLAG	Int	bitwise FLAG
// 3	RNAME	String	References sequence NAME
// 4	POS	Int	1- based leftmost mapping POSition
// 5	MAPQ	Int	MAPping Quality
// 6	CIGAR	String	CIGAR string
// 7	RNEXT	String	Ref. name of the mate/next read
// 8	PNEXT	Int	Position of the mate/next read
// 9	TLEN	Int	observed Template LENgth
// 10	SEQ	String	segment SEQuence
// 11	QUAL	String	ASCII of Phred-scaled base QUALity+33
//
