package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatal(err.Error())
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer conn.Close()
	fmt.Printf("Connection to %s established from %s\n", conn.RemoteAddr(), conn.LocalAddr())

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")

		line, err := reader.ReadString('\n')
		if err != nil {
			log.Print(err)
		}
		_, err = conn.Write([]byte(line))
		if err != nil {
			log.Print(err)
		}

	}
}
