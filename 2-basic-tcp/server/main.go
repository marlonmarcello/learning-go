package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}

	defer ln.Close()

	for {
		// reader and writer interfaces very important
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go handleConnection(conn)

	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// crea a new reader from the connection
	reader := bufio.NewReader(conn)

	// read the command line from the client
	line, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(conn, "Error reading command: %v\n", err)
		return
	}

	// trim the newline character and split the line into command and resource
	parts := strings.SplitN(strings.TrimSpace(line), " ", 2)
	if len(parts) != 2 {
		fmt.Fprintf(conn, "Invalid command format. Expected format is COMMAND:RESOURCE\n")
		return
	}

	command := parts[0]
	resource := parts[1]

	log.Printf("Received command: %s %s\n", command, resource)

	// handle command
	switch command {
	case "GET":
		resolved, err := handleGet(conn, resource)
		if err != nil {
			fmt.Fprintf(conn, "Error on get %s\n%s", resource, err)
		}
		fmt.Println(resolved)
	default:
		fmt.Fprintf(conn, "Unknown command: %s\n", command)
	}
}

type UnkownResourceError struct {
	ResourceName string
}

func (s *UnkownResourceError) Error() string {
	return fmt.Sprintf("Unkown resource %s\n", s.ResourceName)
}

func handleGet(conn net.Conn, resource string) (string, error) {
	// implement GET command handling logic
	fmt.Fprintf(conn, "GET command received for resource: %s\n", resource)

	if strings.Trim(resource, " ") != "/index.html" {
		return "", &UnkownResourceError{resource}
	}

	return "resolved", nil
}
