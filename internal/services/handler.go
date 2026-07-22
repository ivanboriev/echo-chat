package services

import (
	"bufio"
	"chat/internal/models"
	"errors"
	"io"
	"log"
	"net"
	"time"
)

func handleClientMessages(client *models.Client, h *Hub) error {

	h.register <- client

	defer func() {
		h.cleanupClient(client)
		log.Printf("клиент %s отключён", client.ID)
	}()

	scanner := bufio.NewScanner(client.Conn)

	for scanner.Scan() {
		line := scanner.Text()
		isCommand := isCommandExists(line)

		if isCommand {
			h.HandleCommand(client, line)
		} else {
			msg := ParseIncomingMessage(line, client.ID)
			h.history.Add(msg)
			h.broadcast <- msg
		}

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
