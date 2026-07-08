package main

import "chat/internal/services"

func main() {
	err := services.StartEchoServer("8080")
	if err != nil {

	}

}
