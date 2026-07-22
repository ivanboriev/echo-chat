package main

import (
	"chat/internal/services"
	"fmt"
	"os"
)

func main() {
	err := services.StartEchoServer("8080")
	if err != nil {
		fmt.Println(fmt.Errorf("%v", err))
		os.Exit(1)
	}
}
