package main

import (
	"bufio"
	"log"
	"net"
)

func serve(conn net.Conn) {
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
			if err := handle(conn, buf); err != nil {
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

		log.Println(c)

		conn.Write([]byte("Hello\n"))
	}
}

func handle(conn net.Conn, buf []byte) error {
	text := string(buf[:len(buf)-1])
	log.Println("word: " + text)
	conn.Write([]byte("1/Hello/World/Bragi/\n"))
	return nil
}

func main() {
	port := "1234"
	l, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to setup TCP server on port %s: %v", port, err)
	}
	defer l.Close()

	log.Printf("Bragi server is running on port %s\n", port)
	for {
		conn, err := l.Accept()
		log.Println("accept connection")
		if err != nil {
			log.Println("Failed to accpet connection:", err)
			continue
		}

		go serve(conn)
	}
}
