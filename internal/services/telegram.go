package service

import (
	"net/http"
	"net/url"
	"os"
	"strings"
)

const telegram_base_url_api = "https://api.telegram.org/bot"

type iTelegramService interface {
	PostMessage(text string) error
}

type telegramService struct {
	chatID    string
	threadID  string
	baseUrl   string
	parseMode string
}

// Initialisation of the Telegram Service
func initTelegram() *telegramService {
	chatID := os.Getenv("TELEGRAM_CHAT_ID")
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	threadID := os.Getenv("TELEGRAM_THREAT_ID")

	return &telegramService{
		chatID:    chatID,
		threadID:  threadID,
		baseUrl:   telegram_base_url_api + token,
		parseMode: "markdown",
	}
}

// Post Telegram Message
func (service *telegramService) PostMessage(text string) error {
	if os.Getenv("TELEGRAM_DISABLE") == "1" {
		return nil
	}

	data := url.Values{}
	data.Set("text", text)
	data.Set("chat_id", service.chatID)
	data.Set("parse_mode", service.parseMode)

	// If thread provided
	if service.threadID != "" {
		data.Set("reply_to_message_id", service.threadID)
	}

	_, err := http.Post(
		service.baseUrl+"/sendMessage",
		"application/x-www-form-urlencoded",
		strings.NewReader(data.Encode()),
	)

	return err
}
