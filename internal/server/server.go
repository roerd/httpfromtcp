package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync/atomic"

	"github.com/roerd/httpfromtcp/internal/request"
	"github.com/roerd/httpfromtcp/internal/response"
)

type Server struct {
	listener net.Listener
	handler  Handler
	isClosed atomic.Bool
}

type Handler func(w *response.Writer, req *request.Request)

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

func (herr *HandlerError) Write(w io.Writer) error {
	err := response.WriteStatusLine(w, herr.StatusCode)
	if err != nil {
		return err
	}
	headers := response.GetDefaultHeaders(len(herr.Message), "text/plain")
	err = response.WriteHeaders(w, headers)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte("\r\n"))
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(herr.Message))
	return err
}

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	server := &Server{
		listener: listener,
		handler:  handler,
	}
	go server.listen()
	return server, nil
}

func (s *Server) Close() error {
	s.isClosed.Store(true)
	return s.listener.Close()
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.isClosed.Load() {
				return
			}
			log.Println("Error accepting connection:", err)
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	log.Printf("handling connection from %s\n", conn.RemoteAddr())

	defer conn.Close()

	request, err := request.RequestFromReader(conn)
	if err != nil {
		hErr := &HandlerError{
			StatusCode: 400,
			Message:    err.Error(),
		}
		hErr.Write(conn)
		return
	}

	writer := response.NewWriter(conn)

	s.handler(writer, request)
}
