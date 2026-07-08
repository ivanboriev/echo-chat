package services

import (
	"bufio"
	"chat/internal/models"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/google/uuid"
)

func GenerateClientID() string {
	return "User_" + uuid.New().String()
}

func FormatMessage(msg models.ChatMessage) string {
	timeStr := msg.Timestamp.Format("15:04:05")

	if msg.MessageType == "user" {
		return fmt.Sprintf("[%s] <%s>: %s", timeStr, msg.ClientID, msg.Content)
	}

	return fmt.Sprintf("[%s] *** %s", timeStr, msg.Content)
}

func ParseIncomingMessage(raw string, senderID string) models.ChatMessage {
	return models.ChatMessage{
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

	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			log.Printf("Ошибка закрытия сервера: %s", err)
		}
	}(listener)

	hub := NewHub()

	go func(h *Hub) {
		hub.Run()
	}(hub)

	fmt.Printf("TCP Chat Server listening on :%s\n", port)
	fmt.Println("Waiting for connections")

	for {
		conn, err := listener.Accept()

		if err != nil {
			return err
		}

		client := hub.setupClientConnection(conn)

		go func() {
			err := handleClientMessages(client, hub)
			if err != nil {
				log.Printf("Ошибка обработки клиентов: %s", err)
			}
		}()

	}

}
func handleClientMessages(client *models.Client, h *Hub) error {

	h.register <- client

	defer func() {
		h.cleanupClient(client)
		log.Printf("клиент %s отключён", client.ID)
	}()

	scanner := bufio.NewScanner(client.Conn)

	for scanner.Scan() {
		line := scanner.Text()
		msg := ParseIncomingMessage(line, client.ID)
		h.broadcast <- msg
		err := client.Conn.SetReadDeadline(time.Now().Add(ConnectionTimeout * time.Second))
		if err != nil {
			log.Printf("Ошибка обновления таймаута клиента: %s - %s ", client.ID, err)
		}
	}

	defer func(Conn net.Conn) {
		err := Conn.Close()
		if err != nil {
			log.Printf("Ошибка закрытия соединения: %s", err)
		}
	}(client.Conn)

	if err := scanner.Err(); err != nil {
		// Если ошибка – это EOF, значит клиент корректно закрыл соединение.
		if errors.Is(err, io.EOF) {
			h.unregister <- client
			log.Printf("Клиент %s отключился штатно", client.ID)
			return nil
		}
		h.unregister <- client
		// Иначе возвращаем ошибку для обработки вызывающим кодом.
		log.Printf("Клиент %s отключился с ошибкой: %v", client.ID, err)
	}
	return nil
}
