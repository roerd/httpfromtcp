package server

import (
	"bytes"
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

type Handler func(w io.Writer, req *request.Request) *HandlerError

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

func (herr *HandlerError) Write(w io.Writer) error {
	err := response.WriteStatusLine(w, herr.StatusCode)
	if err != nil {
		return err
	}
	headers := response.GetDefaultHeaders(len(herr.Message))
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

	buf := new(bytes.Buffer)

	herr := s.handler(buf, request)
	if herr.StatusCode >= 400 {
		herr.Write(conn)
		return
	}

	body := buf.Bytes()

	err = response.WriteStatusLine(conn, herr.StatusCode)
	if err != nil {
		log.Println("Error writing status line:", err)
		return
	}
	headers := response.GetDefaultHeaders(len(body))
	err = response.WriteHeaders(conn, headers)
	if err != nil {
		log.Println("Error writing headers:", err)
		return
	}
	conn.Write([]byte("\r\n"))
	conn.Write(body)
}
