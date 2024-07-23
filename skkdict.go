package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

var charCodePattern *regexp.Regexp
var concatPattern *regexp.Regexp

func init() {
	charCodePattern = regexp.MustCompile(`\\(0\d+)`)
	concatPattern = regexp.MustCompile(`\(concat\s+"((?:\\.|[^\\"])*)"\)`)
}

type Word struct {
	Text string
	Desc string
}

func (w Word) String() string {
	return w.Text + ";" + w.Desc
}

type Entry struct {
	Label string
	Words []Word
}

type Reader struct {
	scanner *bufio.Scanner
}

type DicMap map[string][]Word

func (r *Reader) Read() (*Entry, error) {
	ok := r.scanner.Scan()
	if !ok {
		return nil, io.EOF
	}

	return parseLine(r.scanner.Text())
}

func (r *Reader) ReadAll() ([]*Entry, error) {
	es := []*Entry{}

	for {
		e, err := r.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return []*Entry{}, errors.WithStack(err)
		}
		if e != nil {
			es = append(es, e)
		}
	}

	return es, nil
}

func (r *Reader) ReadMap() (DicMap, error) {
	mws := DicMap{}

	for {
		e, err := r.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, errors.WithStack(err)
		}
		if e != nil {
			mws[e.Label] = e.Words
		}
	}

	return mws, nil
}

func NewReader(r io.Reader) *Reader {
	sc := bufio.NewScanner(r)

	return &Reader{scanner: sc}
}

func decode(str string) string {
	matches := concatPattern.FindStringSubmatch(str)
	if len(matches) == 0 {
		return str
	}
	str = charCodePattern.ReplaceAllStringFunc(matches[1], func(m string) string {
		s := charCodePattern.FindStringSubmatch(m)
		if len(s) > 1 {
			code, err := strconv.Atoi(s[1])
			if err != nil {
				log.Println(err)
				return m
			}
			return string(rune(code))
		}
		return m
	})
	return str
}

func parseLine(text string) (*Entry, error) {
	if text == "" {
		return nil, nil // empty
	}
	if strings.HasPrefix(text, ";;") {
		return nil, nil // comment
	}
	items := strings.SplitN(text, " ", 2)
	if items == nil || len(items) != 2 {
		return nil, fmt.Errorf("invalid format: %s", text)
	}

	ws := strings.Split(strings.Trim(items[1], "/"), "/")
	words := []Word{}

	for _, w := range ws {
		v := strings.Split(w, ";")
		desc := ""
		if len(v) > 1 {
			desc = decode(v[1])
		}
		words = append(words, Word{
			Text: decode(v[0]),
			Desc: desc,
		})
	}

	return &Entry{Label: items[0], Words: words}, nil
}
