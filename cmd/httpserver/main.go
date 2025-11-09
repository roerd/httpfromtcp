package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/roerd/httpfromtcp/internal/request"
	"github.com/roerd/httpfromtcp/internal/response"
	"github.com/roerd/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(w *response.Writer, req *request.Request) {
	if req.RequestLine.RequestTarget == "/yourproblem" {
		err := w.WriteStatusLine(400)
		if err != nil {
			log.Panicf("Error writing status line: %v", err)
		}
		body := []byte(`<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`)
		headers := response.GetDefaultHeaders(len(body), "text/html")
		err = w.WriteHeaders(headers)
		if err != nil {
			log.Panicf("Error writing headers: %v", err)
		}
		_, err = w.WriteBody(body)
		if err != nil {
			log.Panicf("Error writing body: %v", err)
		}
		return
	}

	if req.RequestLine.RequestTarget == "/myproblem" {
		err := w.WriteStatusLine(500)
		if err != nil {
			log.Panicf("Error writing status line: %v", err)
		}
		body := []byte(`<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`)
		headers := response.GetDefaultHeaders(len(body), "text/html")
		err = w.WriteHeaders(headers)
		if err != nil {
			log.Panicf("Error writing headers: %v", err)
		}
		_, err = w.WriteBody(body)
		if err != nil {
			log.Panicf("Error writing body: %v", err)
		}
		return
	}

	err := w.WriteStatusLine(200)
	if err != nil {
		log.Panicf("Error writing status line: %v", err)
	}
	body := []byte(`<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`)
	headers := response.GetDefaultHeaders(len(body), "text/html")
	err = w.WriteHeaders(headers)
	if err != nil {
		log.Panicf("Error writing headers: %v", err)
	}
	_, err = w.WriteBody(body)
	if err != nil {
		log.Panicf("Error writing body: %v", err)
	}
}
