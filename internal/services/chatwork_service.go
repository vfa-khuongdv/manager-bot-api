package services

import (
	"fmt"
	"io"
	"net/http"
	"net/url" // ← import this
	"strings"
	"time"
)

type IChatworkService interface {
	SendMessage(apiKey, roomId, message string) error
}

type ChatworkService struct {
	BaseURL string
}

func NewChatworkService() *ChatworkService {
	return &ChatworkService{
		BaseURL: "https://api.chatwork.com/v2",
	}
}

func (c *ChatworkService) SendMessage(apiKey, roomId, message string) error {
	// rename local var so it doesn't clash with the url package
	endpoint := fmt.Sprintf("%s/rooms/%s/messages", c.BaseURL, roomId)

	// Prepare URL-encoded form data
	form := url.Values{}
	form.Set("body", message)

	// Create request with encoded form
	req, err := http.NewRequest("POST", endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("X-ChatWorkToken", apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Debug: read response body
	body, _ := io.ReadAll(resp.Body)
	fmt.Println("Response Body:", string(body))

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to send message: received status %d", resp.StatusCode)
	}

	return nil
}
