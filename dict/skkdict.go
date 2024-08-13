package dict

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
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

func isGzip(r io.ReadSeeker) (bool, error) {
	pos, err := r.Seek(0, io.SeekCurrent)
	if err != nil {
		return false, errors.WithStack(err)
	}

	header := make([]byte, 2)
	if _, err := r.Read(header); err != nil {
		return false, errors.WithStack(err)
	}

	if _, err := r.Seek(pos, io.SeekStart); err != nil {
		return false, errors.WithStack(err)
	}

	return header[0] == 0x1f && header[1] == 0x8b, nil
}

func detectEncoding(r io.Reader) encoding.Encoding {
	buf := make([]byte, 512)
	n, err := r.Read(buf)
	if err != nil && err != io.EOF {
		log.Println("Error reading file:", err)
		return unicode.UTF8
	}
	buf = buf[:n]

	if isEUCJP(buf) {
		return japanese.EUCJP
	}

	return unicode.UTF8
}

func NewReader(file *os.File) (*Reader, error) {
	gz, err := isGzip(file)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var r io.Reader = file

	if gz {
		gr, err := gzip.NewReader(file)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		defer gr.Close()
		r = gr
	}

	enc := detectEncoding(r)
	if enc == japanese.EUCJP {
		r = transform.NewReader(r, enc.NewDecoder())
	}

	sc := bufio.NewScanner(r)

	return &Reader{scanner: sc}, nil
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
