package services

import (
	"chat/internal/models"
	"fmt"
	"log"
	"net"
	"time"
)

const (
	ActionListClients = "list_clients"
	ActionClientCount = "client_count"
	ConnectionTimeout = 30
)

type HubRequest struct {
	Action   string
	Response chan interface{}
}

type Hub struct {
	clients    map[string]*models.Client
	broadcast  chan models.ChatMessage
	register   chan *models.Client
	unregister chan *models.Client
	requests   chan HubRequest
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*models.Client),
		broadcast:  make(chan models.ChatMessage),
		register:   make(chan *models.Client),
		unregister: make(chan *models.Client),
		requests:   make(chan HubRequest),
	}
}

func (h *Hub) BroadcastMessage(msg models.ChatMessage) {
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

func (h *Hub) SendUserList(client *models.Client) {
	ac := h.GetActiveClients()
	activeClientsCount := len(ac)
	msg := fmt.Sprintf("Online users (%d): ", activeClientsCount)

	for i, c := range ac {
		if i == activeClientsCount-1 {
			msg = msg + c
		} else {
			msg = msg + c + ", "
		}
	}

	formatted := FormatMessage(models.ChatMessage{
		Timestamp:   time.Time{},
		ClientID:    "system",
		Content:     msg,
		MessageType: "system",
	})

	_, err := client.Conn.Write([]byte(formatted + "\n"))
	if err != nil {
		log.Printf("ошибка записи клиенту %s: %v", client.ID, err)
	}
}

func (h *Hub) HandleCommand(client *models.Client, command string) {
	switch command {
	case "/users":
		h.SendUserList(client)
	case "/help":
		_, err := client.Conn.Write([]byte("Команды: /help, /users, /time, /quit \n"))
		if err != nil {
			log.Printf("ошибка записи клиенту %s: %v", client.ID, err)
		}
	case "/time":
		_, err := client.Conn.Write([]byte(time.Now().String() + "\n"))
		if err != nil {
			log.Printf("ошибка записи клиенту %s: %v", client.ID, err)
		}
	case "/quit":
		h.cleanupClient(client)

	}
}

func (h *Hub) setupClientConnection(conn net.Conn) *models.Client {
	client := &models.Client{
		ID:       GenerateClientID(),
		Conn:     conn,
		JoinTime: time.Now(),
	}
	err := client.Conn.SetReadDeadline(time.Now().Add(ConnectionTimeout * time.Second))
	if err != nil {
		return nil
	}
	_, err = client.Conn.Write([]byte("Hello " + client.ID))

	if err != nil {
		return nil
	}
	return client
}

func (h *Hub) cleanupClient(client *models.Client) {
	err := client.Conn.Close()
	if err != nil {
		return
	}
	h.unregister <- client
	h.broadcast <- models.ChatMessage{
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
