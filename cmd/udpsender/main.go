

package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatalln(err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print(">")
		s, err := reader.ReadString('\n')
		if err != nil {
			log.Println(err)
			continue
		}
		_, err = conn.Write([]byte(s))
		if err != nil {
			log.Println(err)
			continue
		}
	}
}
