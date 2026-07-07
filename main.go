package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/google/uuid"
)

type ChatMessage struct {
	Timestamp   time.Time
	ClientID    string
	Content     string
	MessageType string
}

type Client struct {
	ID       string
	Conn     net.Conn
	JoinTime time.Time
}

func main() {
	StartEchoServer("8080")
}

func GenerateClientID() string {
	return "User_" + uuid.New().String()
}

func HandleClient(client *Client) error {
	scanner := bufio.NewScanner(client.Conn)

	for scanner.Scan() {
		line := scanner.Text()
		msg := ChatMessage{
			Timestamp:   time.Now(),
			ClientID:    client.ID,
			Content:     line,
			MessageType: "user",
		}

		formatedd := FormatMessage(msg)

		fmt.Println(formatedd)
	}

	defer client.Conn.Close()

	if err := scanner.Err(); err != nil {
		// Если ошибка – это EOF, значит клиент корректно закрыл соединение.
		if errors.Is(err, io.EOF) {
			log.Printf("Клиент %s отключился штатно", client.ID)
			return nil
		}
		// Иначе возвращаем ошибку для обработки вызывающим кодом.
		log.Printf("Клиент %s отключился с ошибкой: %v", client.ID, err)
	}
	return nil
}

func FormatMessage(msg ChatMessage) string {
	timeStr := msg.Timestamp.Format("15:04:05")

	if msg.MessageType == "user" {
		return fmt.Sprintf("[%s] <%s>: %s", timeStr, msg.ClientID, msg.Content)
	}

	return fmt.Sprintf("[%s] *** %s", timeStr, msg.Content)
}

func ParseIncomingMessage(raw string, senderID string) ChatMessage {
	return ChatMessage{
		Timestamp:   time.Now(),
		ClientID:    senderID,
		Content:     raw,
		MessageType: "user",
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

	for {
		conn, err := listener.Accept()

		if err != nil {
			return err
		}

		client := &Client{
			ID:       GenerateClientID(),
			Conn:     conn,
			JoinTime: time.Now(),
		}

		go HandleClient(client)

	}

}
