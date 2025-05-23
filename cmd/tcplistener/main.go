

package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
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
		for line := range getLinesChannel(conn) {
			fmt.Println(line)
		}
		fmt.Println("Connection closed")
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	lines := make(chan string)

	go func() {
		b := make([]byte, 8)
		current_line := ""
		
		for {
			n, err := f.Read(b)
			if err != nil {
				break
			}
			parts := strings.Split(string(b[:n]), "\n")
			last := len(parts) - 1
			for _, part := range parts[:last] {
				lines <- current_line + part
				current_line = ""
			}
			current_line += parts[last]
		}
		
		if current_line != "" {
			lines <- current_line
		}

		close(lines)
		f.Close()
	}()

	return lines
}
