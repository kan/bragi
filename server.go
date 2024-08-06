package main

import (
	"bufio"
	"context"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

type Server struct {
	Config *Config
	Dics   []DicMap
}

func LoadServer(config *Config) (*Server, error) {
	log.Printf("Bragi server is running on port %s\n", config.Port)

	if config.UseAI {
		log.Printf("Use AI Dictionary\n")
	}

	dir := config.DictPath
	if dir == "" {
		cdir, err := os.UserCacheDir()
		if err != nil {
			return nil, errors.WithStack(err)
		}
		dir = filepath.Join(cdir, "bragi", "dict")
	}
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, errors.WithStack(err)
	}

	dics := []DicMap{}
	for _, dic := range config.Dictionary {
		m, err := LoadMap(dic, dir)
		if err != nil {
			return nil, err
		}
		log.Printf("Load dictionary: %s ...\n", dic)

		dics = append(dics, m)
	}

	s := &Server{Config: config, Dics: dics}

	return s, nil
}

func (s *Server) Serve(conn net.Conn) {
	defer conn.Close()
	r := bufio.NewReader(conn)
	for {
		c, err := r.ReadByte()
		if err != nil {
			log.Println("Error reading from connection: ", err)
			return
		}
		switch c {
		case '0':
			return
		case '1':
			buf, err := r.ReadBytes(' ')
			if err != nil {
				return
			}
			if err := s.handle(conn, buf); err != nil {
				return
			}
		case '2':
			conn.Write([]byte("0.0.1 "))
		case '3':
			addr := conn.LocalAddr().String()
			conn.Write([]byte(addr + " "))
		default:
			log.Print(c)
		}
	}
}

func (s *Server) handle(conn net.Conn, buf []byte) error {
	text := string(buf[:len(buf)-1])
	log.Println("word: " + text)

	words := []string{}
	if s.Config.UseAI {
		var err error
		words, err = convertKanjiAI(context.Background(), text)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	for _, dic := range s.Dics {
		ws, ok := dic[text]
		if ok {
			for _, w := range ws {
				words = append(words, w.String())
			}
		}
	}

	log.Printf("kanji: %v", words)
	conn.Write([]byte("1/" + strings.Join(words, "/") + "/\n"))
	return nil
}
