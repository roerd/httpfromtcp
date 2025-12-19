package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
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

	if path, ok := strings.CutPrefix(req.RequestLine.RequestTarget, "/httpbin/"); ok {
		err := w.WriteStatusLine(200)
		if err != nil {
			log.Panicf("Error writing status line: %v", err)
		}
		headers := response.GetDefaultHeaders(0, "application/json")
		headers.Delete("Content-Length")
		headers.Set("Transfer-Encoding", "chunked")
		headers.Set("Trailers", "X-Content-SHA256,X-Content-Length")
		err = w.WriteHeaders(headers)
		if err != nil {
			log.Panicf("Error writing headers: %v", err)
		}
		resp, err := http.Get("https://httpbin.org/" + path)
		if err != nil {
			log.Panicf("Error making HTTP request: %v", err)
		}
		defer resp.Body.Close()

		buf := make([]byte, 1024)
		fullBody := make([]byte, 0)
		for {
			n, err := resp.Body.Read(buf)
			if err != nil && err != io.EOF {
				log.Panicf("Error reading response body: %v", err)
			}
			log.Printf("received %d bytes\n", n)
			if n == 0 {
				break
			}
			fullBody = append(fullBody, buf[:n]...)
			_, err = w.WriteChunkedBody(buf[:n])
			if err != nil {
				log.Panicf("Error writing chunk: %v", err)
			}
		}
		_, err = w.WriteChunkedBodyDone()
		if err != nil {
			log.Panicf("Error writing chunk: %v", err)
		}
		hash := sha256.Sum256(fullBody)
		trailers := response.GetNewHeaders()
		trailers.Set("X-Content-SHA256", fmt.Sprintf("%x", hash))
		trailers.Set("X-Content-Length", fmt.Sprintf("%d", len(fullBody)))
		err = w.WriteTrailers(trailers)
		if err != nil {
			log.Panicf("Error writing trailers: %v", err)
		}
		return
	}

	if req.RequestLine.RequestTarget == "/video" {
		err := w.WriteStatusLine(200)
		if err != nil {
			log.Panicf("Error writing status line: %v", err)
		}
		body, err := os.ReadFile("assets/vim.mp4")
		if err != nil {
			log.Panicf("Error reading file: %v", err)
		}
		headers := response.GetDefaultHeaders(len(body), "video/mp4")
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
