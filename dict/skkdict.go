package dict

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
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

type SkkDict struct {
	dictMap map[string][]Word
}

func (d *SkkDict) Convert(word string) ([]string, error) {
	ws, ok := d.dictMap[word]
	if !ok {
		return []string{}, nil
	}

	words := make([]string, len(ws))
	for i, w := range ws {
		words[i] = w.String()
	}

	return words, nil
}

func isURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

func isEUCJP(data []byte) bool {
	for i := 0; i < len(data); i++ {
		if data[i] >= 0x80 {
			if data[i] >= 0xA1 && data[i] <= 0xFE {
				if i+1 >= len(data) {
					return false
				}
				if data[i+1] < 0xA1 || data[i+1] > 0xFE {
					return false
				}
				i++
			} else {
				return false
			}
		}
	}
	return true
}

func NewSkkDict(src, dir string, update bool) (*SkkDict, bool, error) {
	var file *os.File

	if isURL(src) {
		u, err := url.Parse(src)
		if err != nil {
			return nil, false, errors.WithStack(err)
		}

		fname := filepath.Base(u.Path)
		fpath := filepath.Join(dir, fname)

		file, err = os.Open(fpath)
		if err == nil {
			// use cache file
			log.Printf("use: %s\n", fpath)
		} else if os.IsNotExist(err) {
			resp, err := http.Get(src)
			if err != nil {
				return nil, false, errors.WithStack(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return nil, false, fmt.Errorf("failed to download dictionary: %s", resp.Status)
			}

			file, err = os.Create(fpath)
			if err != nil {
				return nil, false, errors.WithStack(err)
			}
			defer file.Close()

			if _, err := io.Copy(file, resp.Body); err != nil {
				return nil, false, errors.WithStack(err)
			}
			if _, err := file.Seek(0, io.SeekStart); err != nil {
				return nil, false, errors.WithStack(err)
			}
		} else {
			return nil, false, errors.WithStack(err)
		}
	} else {
		var err error
		file, err := os.Open(src)
		if err != nil {
			return nil, false, errors.WithStack(err)
		}
		defer file.Close()
		update = false // ローカル辞書は更新しない
	}

	r, err := NewReader(file)
	if err != nil {
		return nil, update, errors.WithStack(err)
	}

	m, err := r.ReadMap()
	sd := &SkkDict{dictMap: m}
	return sd, update, err
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
