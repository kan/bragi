package server

import (
	"bufio"
	"context"
	"log"
	"net"
	"strings"

	"github.com/kan/bragi/config"
	"github.com/kan/bragi/dict"
	"github.com/kan/bragi/openai"
	"github.com/pkg/errors"
)

type Server struct {
	Config *config.Config
	Dics   []dict.DicMap
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
	if len(text) > 1 && s.Config.UseAI {
		var err error
		words, err = openai.ConvertKanjiAI(context.Background(), text)
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

func LoadServer(conf *config.Config) (*Server, error) {
	log.Printf("Bragi server is running on port %s\n", conf.Port)

	if conf.UseAI {
		log.Printf("Use AI Dictionary\n")
	}

	dir, err := conf.GetCacheDir()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	dics := []dict.DicMap{}
	for _, dic := range conf.Dictionary {
		m, _, err := dict.LoadMap(dic, dir, false)
		if err != nil {
			return nil, err
		}
		log.Printf("Load dictionary: %s ...\n", dic)

		dics = append(dics, m)
	}

	s := &Server{Config: conf, Dics: dics}

	return s, nil
}
