package server

import (
	"bufio"
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
	Dicts  []dict.Dict
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
	for _, dic := range s.Dicts {
		ws, err := dic.Convert(text)
		if err == nil {
			words = append(words, ws...)
		}
	}

	log.Printf("kanji: %v", words)
	conn.Write([]byte("1/" + strings.Join(words, "/") + "/\n"))
	return nil
}

func LoadServer(conf *config.Config) (*Server, error) {
	log.Printf("Bragi server is running on port %s\n", conf.Port)

	dics := []dict.Dict{}
	if conf.UseAI {
		ad := openai.NewOpenAIDict()
		dics = append(dics, ad)
		log.Printf("Use AI Dictionary\n")
	}
	if conf.UseLisp {
		ld := dict.NewLispDict()
		dics = append(dics, ld)
		log.Printf("Use Lisp Dictionary\n")
	}

	dir, err := conf.GetCacheDir()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	for _, dic := range conf.Dictionary {
		sd, _, err := dict.NewSkkDict(dic, dir, false)
		if err != nil {
			return nil, err
		}
		log.Printf("Load dictionary: %s ...\n", dic)

		dics = append(dics, sd)
	}

	s := &Server{Config: conf, Dicts: dics}

	return s, nil
}
