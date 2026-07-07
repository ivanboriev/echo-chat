package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"time"
)

type ChatMessage struct {
	Timestamp    time.Time
	ClientID     string
	Content      string
	MessaageType string
}

func main() {
	StartEchoServer("8080")
}

func FormatMessage(msg ChatMessage) string {
	timeStr := msg.Timestamp.Format("15:04:05")

	if msg.MessaageType == "user" {
		return fmt.Sprintf("[%s] <%s>: %s", timeStr, msg.ClientID, msg.Content)
	}

	return fmt.Sprintf("[%s] *** %s", timeStr, msg.Content)

}

func ParseIncomingMessage(raw string, senderID string) ChatMessage {
	return ChatMessage{
		Timestamp:    time.Now(),
		ClientID:     senderID,
		Content:      raw,
		MessaageType: "user",
	}

}

func StartEchoServer(port string) error {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	defer listener.Close()

	fmt.Printf("TCP Echo Server listening on :%s\n", port)
	fmt.Println("Waiting for connections")

	conn, err := listener.Accept()

	if err != nil {
		return err
	}

	defer conn.Close()

	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		msg := ParseIncomingMessage(scanner.Text(), "Ivan Boriev")
		conn.Write([]byte(formatMessage(msg) + "\n"))
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("scanning failed: %v", err)
	}

	return nil
}
