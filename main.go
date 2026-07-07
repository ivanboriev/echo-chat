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

type Hub struct {
	clients    map[string]*Client
	broadcast  chan ChatMessage
	register   chan *Client
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		broadcast:  make(chan ChatMessage),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) BroadcastMessage(msg ChatMessage) {
	formatted := FormatMessage(msg)

	for id, client := range h.clients {

		if id == msg.ClientID {
			continue
		}

		_, err := client.Conn.Write([]byte(formatted + "\n"))
		if err != nil {
			log.Printf("ошибка записи клиенту %s: %v", id, err)
			select {
			case h.unregister <- client:
			default:
				log.Printf("не удалось отправить клиента %s в unregister (канал занят)", id)
			}
		}
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client.ID] = client
			log.Printf("клиент %s зарегистрирован", client.ID)

		case client := <-h.unregister:
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				log.Printf("клиент %s удалён", client.ID)
			}

		case msg := <-h.broadcast:
			h.BroadcastMessage(msg)
		}
	}
}

func main() {
	StartEchoServer("8080")
}

func GenerateClientID() string {
	return "User_" + uuid.New().String()
}

func HandleClient(client *Client, h *Hub) error {

	h.register <- client

	defer func() {
		h.unregister <- client
		log.Printf("клиент %s отключён", client.ID)
	}()

	scanner := bufio.NewScanner(client.Conn)

	for scanner.Scan() {
		line := scanner.Text()
		msg := ChatMessage{
			Timestamp:   time.Now(),
			ClientID:    client.ID,
			Content:     line,
			MessageType: "user",
		}

		h.broadcast <- msg
	}

	defer client.Conn.Close()

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

	hub := NewHub()
	go hub.Run()

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

		go HandleClient(client, hub)

	}

}
