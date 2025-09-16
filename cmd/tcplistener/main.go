package main

import (
	"fmt"
	"httpfromtcp/internal/request"
	"log"
	"net"
)

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

		request, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal(err)
		}

		requestLine := fmt.Sprintf("Request line:\n- Method: %s\n- Target: %s\n- Version: %s\n Headers:", request.RequestLine.Method, request.RequestLine.HTTPVersion, request.RequestLine.RequestTarget)

		for key, value := range request.Headers {
			requestLine += fmt.Sprintf("\n- %s: %s", key, value)
		}
		fmt.Println(requestLine)
		conn.Close()
	}
}
