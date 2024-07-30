package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

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

func LoadMap(src string) (DicMap, error) {
	var reader io.Reader

	if isURL(src) {
		resp, err := http.Get(src)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to download dictionary: %s", resp.Status)
		}

		u, err := url.Parse(src)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		fname := filepath.Base(u.Path)

		file, err := os.Create(fname)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		defer file.Close()

		if _, err := io.Copy(file, resp.Body); err != nil {
			return nil, errors.WithStack(err)
		}
		if _, err := file.Seek(0, io.SeekStart); err != nil {
			return nil, errors.WithStack(err)
		}

		reader = file
	} else {
		var err error
		file, err := os.Open(src)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		defer file.Close()

		reader = file
	}

	if strings.HasSuffix(src, ".gz") {
		gr, err := gzip.NewReader(reader)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		defer gr.Close()

		reader = gr
	}

	enc := detectEncoding(reader)
	if enc == japanese.EUCJP {
		reader = transform.NewReader(reader, enc.NewDecoder())
	}

	r := NewReader(reader)
	return r.ReadMap()
}
