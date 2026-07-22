package services

import (
	"fmt"
	"log"
	"net"
)

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
