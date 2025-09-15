package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	chanel := make(chan string)

	go func() {
		defer close(chanel)
		defer f.Close()
		fullStr := ""
		for {
			buffer := make([]byte, 8)
			bytesLen, err := f.Read(buffer)
			if err != nil {
				if fullStr != "" {
					chanel <- fullStr
				}
				if errors.Is(err, io.EOF) {
					break
				} else {
					log.Fatal(err)
				}
			}

			fullStr += string(buffer[:bytesLen])

			newLineStrings := strings.Split(fullStr, "\n")
			for i := 0; i < len(newLineStrings)-1; i++ {
				chanel <- newLineStrings[i]
			}
			fullStr = newLineStrings[len(newLineStrings)-1]
		}
	}()

	return chanel
}

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Connection accepted from, %s\n", conn.RemoteAddr())

		c := getLinesChannel(conn)
		for line := range c {
			fmt.Printf("%s\n", line)
		}
		fmt.Printf("Connection from %s closed\n", conn.RemoteAddr())
	}
}
