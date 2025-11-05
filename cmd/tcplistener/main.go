package main

import (
	"fmt"
	"log"
	"net"

	"github.com/roerd/httpfromtcp/internal/request"
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
		fmt.Println("Connection accepted")
		request, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal(err)
		}

		requestLine := request.RequestLine
		fmt.Println("Request line:")
		fmt.Println("- Method:", requestLine.Method)
		fmt.Println("- Target:", requestLine.RequestTarget)
		fmt.Println("- Version:", requestLine.HttpVersion)

		fmt.Println("Headers:")
		for key, value := range request.Headers {
			fmt.Printf("- %s: %s\n", key, value)
		}

		fmt.Println("Body:")
		fmt.Println(string(request.Body))

		fmt.Println("Connection closed")
	}
}
