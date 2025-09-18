package server

import (
	"fmt"
	"net"
	"sync/atomic"
)

var stdResponse = fmt.Sprint(
	"HTTP/1.1 200 OK\r\n" +
		"Content-Type: text/plain\n" +
		"Content-Length: 13\n" +
		"\r\n\r\n" +
		"Hello World!\n",
)

type Server struct {
	Listener net.Listener
	State    atomic.Bool
}

func newServer() Server {
	return Server{}
}

func (s *Server) Close() error {
	err := s.Listener.Close()
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) listen() {
	s.State.Store(true)
	for {
		connection, err := s.Listener.Accept()
		if !s.State.Load() {
			break
		}
		if err != nil {
			panic("you are a bad progremmer")
		}
		s.handle(connection)
	}
}

func (s *Server) handle(conn net.Conn) {
	conn.Write([]byte(stdResponse))
}

func Serve(port int) (*Server, error) {
	newServer := newServer()

	newListener, _ := net.Listen("tcp", ":"+fmt.Sprint(port))
	newServer.Listener = newListener
	go func() {
		newServer.listen()
	}()

	return &newServer, nil
}
