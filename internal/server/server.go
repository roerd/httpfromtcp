package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/roerd/httpfromtcp/internal/response"
)

type Server struct {
	listener net.Listener
	isClosed atomic.Bool
}

func Serve(port int) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	server := &Server{listener: listener}
	go server.listen()
	return server, nil
}

func (s *Server) Close() error {
	s.isClosed.Store(true)
	return s.listener.Close()
}

func (s *Server) listen() {
	for {
		if s.isClosed.Load() {
			return
		}
		conn, err := s.listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	log.Printf("handling connection from %s\n", conn.RemoteAddr())

	defer conn.Close()

	body := []byte("")

	err := response.WriteStatusLine(conn, 200)
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
