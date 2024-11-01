package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
	// connects to server
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	// write to server
	fmt.Fprintf(conn, "GET /index.html\n")

	// read the response from the server
	bs := make([]byte, 1024)
	n, err := conn.Read(bs)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(bs[:n]))
}
