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

type HubRequest struct {
	Action   string
	Response chan interface{}
}

const (
	ActionListClients = "list_clients"
	ActionClientCount = "client_count"
	ConnectionTimeout = 5
)

type Hub struct {
	clients    map[string]*Client
	broadcast  chan ChatMessage
	register   chan *Client
	unregister chan *Client
	requests   chan HubRequest
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		broadcast:  make(chan ChatMessage),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		requests:   make(chan HubRequest),
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
func (h *Hub) GetActiveClients() []string {
	resp := make(chan interface{})
	h.requests <- HubRequest{Action: ActionListClients, Response: resp}
	return (<-resp).([]string)
}

func (h *Hub) GetClientCount() int {
	resp := make(chan interface{})
	h.requests <- HubRequest{Action: ActionClientCount, Response: resp}
	return (<-resp).(int)
}

func (h *Hub) setupClientConnection(conn net.Conn) *Client {
	client := &Client{
		ID:       GenerateClientID(),
		Conn:     conn,
		JoinTime: time.Now(),
	}
	err := client.Conn.SetReadDeadline(time.Now().Add(ConnectionTimeout * time.Second))
	if err != nil {
		return nil
	}
	_, err = client.Conn.Write([]byte("Hello" + client.ID))

	if err != nil {
		return nil
	}
	return client
}

func (h *Hub) cleanupClient(client *Client) {
	err := client.Conn.Close()
	if err != nil {
		return
	}
	h.unregister <- client
	h.broadcast <- ChatMessage{
		Timestamp:   time.Now(),
		ClientID:    client.ID,
		Content:     fmt.Sprintf("Клиен %s отключился от сервера", client.ID),
		MessageType: "system",
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client.ID] = client
			log.Printf("Client %s connected. Total: %d", client.ID, len(h.clients))

		case client := <-h.unregister:
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				log.Printf("Client %s disconnected. Total: %d", client.ID, len(h.clients))
			}

		case msg := <-h.broadcast:
			h.BroadcastMessage(msg)

		case req := <-h.requests:
			switch req.Action {
			case ActionListClients:
				var clients []string
				for id := range h.clients {
					clients = append(clients, id)
				}
				req.Response <- clients
			case ActionClientCount:
				req.Response <- len(h.clients)
			}
		}
	}
}

func GenerateClientID() string {
	return "User_" + uuid.New().String()
}

func handleClientMessages(client *Client, h *Hub) error {

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

	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			log.Printf("Ошибка закрытия сервера: %s", err)
		}
	}(listener)

	hub := NewHub()
	go hub.Run()

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

func main() {
	err := StartEchoServer("8080")
	if err != nil {

	}

}
