package fastats

import (
	"bufio"
	"io"
	"unicode"
	"strings"
	"strconv"
	"fmt"
)

type token interface {
	isToken()
}

type lparen struct{}
func (t lparen) isToken(){}

type rparen struct{}
func (t rparen) isToken(){}

type comma struct{}
func (t comma) isToken(){}

type semicolon struct{}
func (t semicolon) isToken(){}

type colon struct{}
func (t colon) isToken(){}

type sstring string
func (t sstring) isToken(){}

type num float64
func (t num) isToken(){}

type Newick struct {
	Name string
	Length float64
	Children []*Newick
}

type runeReadUnreader interface {
	io.RuneReader
	UnreadRune() error
}

type tokenReader struct {
	r runeReadUnreader
	hold token
	full bool
}

func (tr *tokenReader) ReadToken() (t token, err error) {
	if tr.full {
		tr.full = false
		return tr.hold, nil
	}
	for {
		r, _, e := tr.r.ReadRune()
		if e != nil {
			return nil, e
		}
		if unicode.IsLetter(r) {
			if e := tr.r.UnreadRune(); e != nil {
				return nil, e
			}
			return tr.tokenizeString()
		}
		if unicode.IsNumber(r) {
			if e := tr.r.UnreadRune(); e != nil {
				return nil, e
			}
			return tr.tokenizeNum()
		}
		switch r {
		case '(': return lparen{}, nil
		case ')': return rparen{}, nil
		case ',': return comma{}, nil
		case ';': return semicolon{}, nil
		case ':': return colon{}, nil
		default:
		}
	}
	return t, nil
}

func (tr *tokenReader) tokenizeString() (sstring, error) {
	var b strings.Builder
	for {
		r, _, e := tr.r.ReadRune()
		if e != nil {
			return sstring(b.String()), io.EOF
		}
		if !unicode.IsLetter(r) {
			e := tr.r.UnreadRune()
			if e != nil {
				return sstring(b.String()), e
			}
			break
		}
		b.WriteRune(r)
	}
	return sstring(b.String()), nil
}

func (tr *tokenReader) tokenizeNumberString() (string, error) {
	var b strings.Builder
	for {
		r, _, e := tr.r.ReadRune()
		if e != nil {
			return b.String(), io.EOF
		}
		if !unicode.IsDigit(r) && r != 'e' && r != 'E' && r != '-' {
			if e := tr.r.UnreadRune(); e != nil {
				return b.String(), e
			}
			break
		}
		b.WriteRune(r)
	}
	return b.String(), nil
}

func (tr *tokenReader) tokenizeNum() (num, error) {
	s, e := tr.tokenizeNumberString()
	if e != nil {
		return num(0), nil
	}
	n, e := strconv.ParseFloat(s, 64)
	return num(n), e
}

func (tr *tokenReader) UnreadToken(t token) {
	if tr.full {
		panic(fmt.Errorf("tokenReader.UnreadToken: tried to put token %v into full token reader", t))
	}
	tr.hold = t
	tr.full = true
}

func parseChildren(tr *tokenReader) ([]*Newick, error) {
	var chs []*Newick
	mainloop: for {
		child, e := parseNewick(tr)
		if e != nil {
			return chs, e
		}
		chs = append(chs, child)

		t, e := tr.ReadToken()
		if e != nil {
			return chs, e
		}
		switch t.(type) {
		case rparen, semicolon:
			break mainloop
		default:
		}
	}
	return chs, nil
}

func parseNewick(tr *tokenReader) (*Newick, error) {
	n := &Newick{}
	for {
		t, e := tr.ReadToken()
		if e != nil {
			return n, e
		}
		switch v := t.(type) {
		case comma, rparen, semicolon:
			tr.UnreadToken(t)
			return n, nil
		case sstring:
			n.Name = string(v)
		case colon:
			t2, e2 := tr.ReadToken()
			if e2 != nil {
				return n, e2
			}
			num, ok := t2.(num)
			if !ok {
				return n, fmt.Errorf("parseNewick: colon followed by non-number %v", t2)
			}
			n.Length = float64(num)
		case lparen:
			n.Children, e = parseChildren(tr)
			if e != nil {
				return n, e
			}
		case num:
		default:
			return n, fmt.Errorf("parseNewick: impossible type for token %v", t)
		}
	}
	return n, nil
}

func ParseNewick(r io.Reader) (*Newick, error) {
	if br, ok := r.(runeReadUnreader); ok {
		n, e := parseNewick(&tokenReader{r: br})
		if e == io.EOF {
			return n, nil
		}
		return n, e
	}
	n, e := parseNewick(&tokenReader{r: bufio.NewReader(r)})
	if e == io.EOF {
		return n, nil
	}
	return n, e
}

func printRep(w io.Writer, s string, n int) (nw int, err error) {
	for i := 0; i < n; i++ {
		nwrit, e := fmt.Fprintf(w, "%v", s)
		nw += nwrit
		if e != nil {
			return nw, e
		}
	}
	return nw, nil
}

func printNewick(w io.Writer, n *Newick, indent int) (nw int, err error) {
	nw, e := printRep(w, "\t", indent)
	if e != nil {
		return nw, e
	}
	nwrit, e := fmt.Fprintf(w, "(name: \"%v\", length: \"%v\"", n.Name, n.Length)
	nw += nwrit
	if e != nil {
		return nw, e
	}
	if len(n.Children) < 1 {
		nwrit, e := fmt.Fprintf(w, ")\n")
		nw += nwrit
		return nw, e
	}
	nwrit, e = fmt.Fprintf(w, ", children:\n")
	nw += nwrit
	if e != nil {
		return nw, e
	}
	for _, child := range n.Children {
		nwrit, e := printNewick(w, child, indent + 1)
		nw += nwrit
		if e != nil {
			return nw, e
		}
	}
	nwrit, e = printRep(w, "\t", indent)
	nw += nwrit
	if e != nil {
		return nw, e
	}
	nwrit, e = fmt.Fprintf(w, ")\n")
	nw += nwrit
	return nw, e
}

func PrintNewickDetailed(w io.Writer, n *Newick) (nw int, err error) {
	return printNewick(w, n, 0)
}

func PrintNewick(w io.Writer, n *Newick) (nw int, err error) {
	if len(n.Children) > 0 {
		nwrit, e := fmt.Fprintf(w, "(")
		nw += nwrit
		if e != nil {
			return nw, e
		}
		for i, child := range n.Children {
			if i > 0 {
				nwrit, e = fmt.Fprintf(w, ",")
				nw += nwrit
				if e != nil {
					return nw, e
				}
			}
			nwrit, e = PrintNewick(w, child)
			nw += nwrit
			if e != nil {
				return nw, e
			}
		}
		nwrit, e = fmt.Fprintf(w, ")")
		nw += nwrit
		if e != nil {
			return nw, e
		}
	}
	if len(n.Name) > 0 {
		nwrit, e := fmt.Fprintf(w, "%s", n.Name)
		nw += nwrit
		if e != nil {
			return nw, e
		}
	}
	if n.Length != 0.0 {
		nwrit, e := fmt.Fprintf(w, ":%g", n.Length)
		nw += nwrit
		if e != nil {
			return nw, e
		}
	}
	return nw, nil
}

func (n *Newick) String() string {
	var b strings.Builder
	PrintNewick(&b, n)
	return b.String()
}
