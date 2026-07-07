package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

func main() {
	StartEchoServer("8080")
}

func StartEchoServer(port string) error {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	fmt.Printf("TCP Echo Server listening on :%s\n", port)
	fmt.Println("Waiting for connections")

	defer listener.Close()

	conn, err := listener.Accept()
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		conn.Write([]byte(scanner.Text() + "\n"))
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("scanning failed: %v", err)
	}

	return nil
}
