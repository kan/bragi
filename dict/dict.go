package dict

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
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

func LoadMap(src, dir string, update bool) (DicMap, bool, error) {
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
	return m, update, err
}
