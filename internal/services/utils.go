package services

import (
	"chat/internal/models"
	"fmt"
	"slices"
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

func isCommandExists(cmd string) bool {
	return slices.Contains([]string{"/help", "/users", "/time", "/quit"}, cmd)
}
