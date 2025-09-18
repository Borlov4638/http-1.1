package server

import (
	"fmt"
	"httpfromtcp/internal/response"
	"net"
	"sync/atomic"
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
	defer conn.Close()
	err := response.WriteStatusLine(conn, response.StatusCodeOk)
	if err != nil {
		fmt.Println(err)
	}
	resHeaders := response.GetDefaultHeaders(0)
	err = response.WriteHeaders(conn, resHeaders)
	if err != nil {
		fmt.Println(err)
	}
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
