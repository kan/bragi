package main

import (
	"log"
	"net"
	"os"
)

func main() {
	cfile := ""
	if len(os.Args) > 1 {
		cfile = os.Args[1]
	}
	config, err := LoadConfig(cfile)
	if err != nil {
		panic(err)
	}

	if len(os.Args) == 3 {
		if os.Args[2] != "update" {
			log.Fatalf("unsupport subcommand '%s'.", os.Args[2])
		}

		dir, err := getCacheDir(config)
		if err != nil {
			panic(err)
		}
		for _, dic := range config.Dictionary {
			_, up, err := LoadMap(dic, dir, false)
			if err != nil {
				panic(err)
			}
			if up {
				log.Printf("Update dictionary: %s ...\n", dic)
			}
		}
		os.Exit(0)
	}

	l, err := net.Listen("tcp", ":"+config.Port)
	if err != nil {
		log.Fatalf("Failed to setup TCP server on port %s: %v", config.Port, err)
	}
	defer l.Close()

	s, err := LoadServer(config)
	if err != nil {
		panic(err)
	}

	for {
		conn, err := l.Accept()
		log.Println("accept connection")
		if err != nil {
			log.Println("Failed to accpet connection:", err)
			continue
		}

		go s.Serve(conn)
	}
}
