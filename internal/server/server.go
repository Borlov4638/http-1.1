package server

import (
	"bytes"
	"fmt"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"io"
	"net"
	"sync/atomic"
)

type Server struct {
	Listener net.Listener
	State    atomic.Bool
	Handler  Handler
}

func newServer(h Handler) Server {
	return Server{
		Handler: h,
	}
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

	request, err := request.RequestFromReader(conn)
	if err != nil {
		fmt.Println(err)
	}

	buffer := bytes.NewBuffer([]byte{})
	handlerErr := s.Handler(buffer, request)
	if handlerErr != nil {
		err = response.WriteStatusLine(conn, handlerErr.StatusCode)
		if err != nil {
			fmt.Println(err)
		}
		return
	}

	err = response.WriteStatusLine(conn, response.StatusCodeOk)
	if err != nil {
		fmt.Println(err)
	}
	resHeaders := response.GetDefaultHeaders(len(buffer.Bytes()))
	err = response.WriteHeaders(conn, resHeaders)
	if err != nil {
		fmt.Println(err)
	}
	conn.Write(buffer.Bytes())
}

func Serve(port int, h Handler) (*Server, error) {
	newServer := newServer(h)

	newListener, _ := net.Listen("tcp", ":"+fmt.Sprint(port))
	newServer.Listener = newListener
	go func() {
		newServer.listen()
	}()

	return &newServer, nil
}

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

func (he *HandlerError) SendError(w io.Writer) {
	errMsg := fmt.Sprintf("HTTP/1.1 %v %s", he.StatusCode, he.Message)
	w.Write([]byte(errMsg))
}

type Handler func(w io.Writer, req *request.Request) *HandlerError
