package fastats

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"bufio"
	"regexp"

	"github.com/jgbaldwinbrown/zfile"
)

func LineCount(r io.Reader) (n int64, err error) {
	br := bufio.NewReader(r)
	for c, e := br.ReadByte(); e != io.EOF; c, e = br.ReadByte() {
		if e != nil {
			return n, e
		}
		if c == '\n' {
			n++
		}
	}
	return n, nil
}

func LineCountPath(path string) (n int64, err error) {
	r, e := zfile.Open(path)
	if e != nil {
		return 0, e
	}
	defer func() {
		e := r.Close()
		if err == nil {
			err = e
		}
	}()
	return LineCount(r)
}

var extRe = regexp.MustCompile(`\.[^/]*$`)
var fqRe = regexp.MustCompile(`^\.f(ast)?q`)

func AppendLines(lines []string, s *bufio.Scanner, n int) ([]string, error) {
	for i := 0; i < n; i++ {
		if !s.Scan() {
			return lines, io.EOF
		}
		if s.Err() != nil {
			return lines, s.Err()
		}
		lines = append(lines, s.Text())
	}
	return lines, nil
}

func WriteLines(w io.Writer, lines []string) error {
	for _, line := range lines {
		if _, e := fmt.Fprintf(w, "%s\n", line); e != nil {
			return e
		}
	}
	return nil
}

func SplitOneFq(path, pathOutdir, base, outsuffix string, entriesPerPiece int64) (err error) {
	r, e := zfile.Open(path)
	if e != nil {
		return e
	}
	defer func() {
		e := r.Close()
		if err == nil {
			err = e
		}
	}()

	s := bufio.NewScanner(r)
	s.Buffer([]byte{}, 1e15)
	lines := make([]string, 0, 4)
	for pathi := 0; ; pathi++ {
		opath := filepath.Join(pathOutdir, fmt.Sprintf("%s_%04d%s", base, pathi, outsuffix))
		w, e := zfile.Create(opath)
		if e != nil {
			return e
		}
		var written int64 = 0
		var i int64
		for i = 0; i < entriesPerPiece; i++ {
			lines, e = AppendLines(lines[:0], s, 4)
			if e == io.EOF {
				w.Close()
				return nil
			}
			if e != nil {
				w.Close()
				return e
			}
			if e := WriteLines(w, lines); e != nil {
				w.Close()
				return e
			}
			written++
		}
		if e := w.Close(); e != nil {
			return e
		}
		if written < 1 {
			if e := os.Remove(opath); e != nil {
				return e
			}
		}
	}
	return nil
}

func SplitFq(pieces int, outdir string, paths ...string) error {
	if len(paths) < 1 {
		return nil
	}
	if outdir != "" {
		if e := os.MkdirAll(outdir, 0755); e != nil {
			return e
		}
	}
	lines, e := LineCountPath(paths[0])
	if e != nil {
		return e
	}
	for _, path := range paths[1:] {
		plines, e := LineCountPath(path)
		if e != nil {
			return e
		}
		if plines != lines {
			return fmt.Errorf("SplitFq: plines %v for path %v != lines %v for path %v", plines, path, lines, paths[0])
		}
	}
	if lines % 4 != 0 {
		return fmt.Errorf("SplitFq: lines % 4 != 0; lines %v; lines % 4 = %v", lines, lines % 4)
	}
	entries := lines / 4
	entriesPerPiece := entries / int64(pieces)
	if entries % int64(pieces) != 0 {
		entriesPerPiece++
	}

	for _, path := range paths {
		bigext := extRe.FindString(path)
		if bigext == "" || !fqRe.MatchString(bigext) {
			return fmt.Errorf("SplitFq: path %v does not have .fq extension", path)
		}
		ext := filepath.Ext(path)
		outsuffix := ".fq"
		if bigext != ext {
			outsuffix = ".fq" + ext
		}
		pathOutdir := outdir
		if outdir == "" {
			pathOutdir = filepath.Dir(path)
		}

		base := filepath.Base(extRe.ReplaceAllString(path, ""))
		if e := SplitOneFq(path, pathOutdir, base, outsuffix, entriesPerPiece); e != nil {
			return e
		}
	}
	return nil
}
