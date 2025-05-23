package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/vfa-khuongdv/golang-cms/internal/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
)

type IHookService interface {
	ChatworkHook(payload DiscordPayload) error
	SlackHook(payload SlackPayload) error
}

type HookService struct {
	cw IChatworkService
}

type DiscordField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

type DiscordFooter struct {
	Text string `json:"text"`
}

type DiscordEmbed struct {
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Fields      []DiscordField `json:"fields"`
	Footer      DiscordFooter  `json:"footer"`
}

type DiscordPayload struct {
	Embeds []DiscordEmbed `json:"embeds"`
}

type SlackPayload struct {
	Text        string       `json:"text"`
	Blocks      []Block      `json:"blocks,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

type Attachment struct {
	Color  string  `json:"color,omitempty"`
	Blocks []Block `json:"blocks,omitempty"`
}

type Block struct {
	Type string    `json:"type"`
	Text *TextData `json:"text,omitempty"`
}

type TextData struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type Info struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

func NewHookService(cw IChatworkService) *HookService {
	return &HookService{
		cw: cw,
	}
}

func (h *HookService) ChatworkHook(payload DiscordPayload) error {
	chatworkString := ConvertDiscordPayloadToChatwork(payload)
	logger.Infof("Converted Chatwork message: %s", chatworkString)

	// Send the message to Chatwork
	ROOM_ID := utils.GetEnv("CHATWORK_ROOM_ID", "")
	API_KEY := utils.GetEnv("CHATWORK_API_TOKEN", "")

	err := h.cw.SendMessage(API_KEY, ROOM_ID, chatworkString)
	if err != nil {
		return err
	}

	return nil
}

func ConvertDiscordPayloadToChatwork(payload DiscordPayload) string {
	var builder strings.Builder

	for _, embed := range payload.Embeds {
		builder.WriteString("[info]")

		// Title with removed emojis replaced
		title := removeDiscordIcons(embed.Title)
		if title != "" {
			builder.WriteString("[title]" + strings.TrimSpace(title) + "[/title]\n")
		}

		// Description
		if embed.Description != "" {
			description := convertMarkdownLinks(embed.Description)
			description = convertDiscordIconsToChatwork(description)
			description = strings.Replace(description, "Open application", "🔗 Application URL", -1)
			builder.WriteString(description + "\n\n")
		}

		// Field section
		for _, field := range embed.Fields {
			value := convertMarkdownLinks(field.Value)
			value = convertTimestamps(value)
			value = convertValueToChatwork(value)

			icon := getFieldIcon(field.Name)
			builder.WriteString(fmt.Sprintf("%s %s: %s\n", icon, field.Name, value))
		}

		// Footer
		if embed.Footer.Text != "" {
			builder.WriteString("\n[hr]\n")
			builder.WriteString("From Discord, sending all our love 💖🤝💬\n")
		}

		builder.WriteString("[/info]\n\n")
	}

	return builder.String()
}

// Add icons to fields
func getFieldIcon(name string) string {
	switch strings.ToLower(name) {
	case "status":
		return "✅"
	case "author":
		return "👤"
	case "project":
		return "📦"
	case "environment":
		return "🌍"
	case "name":
		return "📛"
	case "deployment logs":
		return "📄"
	case "time":
		return "🕒"
	default:
		return "📊"
	}
}

// move emoji Discord to icon Chatwork
func convertDiscordIconsToChatwork(input string) string {
	replacer := strings.NewReplacer(
		":white_check_mark:", "✅",
		":x:", "❌",
		":warning:", "⚠️",
		":information_source:", "ℹ️",
		":cross_mark:", "❌",
	)
	return replacer.Replace(input)
}

// convert Discord message to Chatwork format
func convertValueToChatwork(input string) string {
	replacer := strings.NewReplacer(
		"Link:", "",
	)
	return replacer.Replace(input)
}

// Remove icon Discord
func removeDiscordIcons(input string) string {
	replacer := strings.NewReplacer(
		":white_check_mark:", "",
		":x:", "",
		":warning:", "",
		":information_source:", "",
		":cross_mark:", "",
	)
	return replacer.Replace(input)
}

// Move markdown [Link](url) → Link: url
func convertMarkdownLinks(input string) string {
	re := regexp.MustCompile(`\[(.*?)\]\((.*?)\)`)
	return re.ReplaceAllString(input, "$1: $2")
}

// Convert Discord timestamps to a readable format
// Example: <t:1234567890:R> → 2006-01-02 15:04:05
func convertTimestamps(input string) string {
	re := regexp.MustCompile(`<t:(\d+):R>`)
	return re.ReplaceAllStringFunc(input, func(m string) string {
		matches := re.FindStringSubmatch(m)
		if len(matches) < 2 {
			return m
		}
		unixTime, err := strconv.ParseInt(matches[1], 10, 64)
		if err != nil {
			return m
		}
		t := time.Unix(unixTime, 0).Local()
		return t.Format("2006-01-02 15:04:05")
	})
}

func (h *HookService) SlackHook(payload SlackPayload) error {
	chatworkString := ConvertSlackPayloadToChatwork(payload)
	logger.Infof("Converted Chatwork message: %s", chatworkString)

	ROOM_ID := utils.GetEnv("CHATWORK_ROOM_ID", "")
	API_KEY := utils.GetEnv("CHATWORK_API_TOKEN", "")

	err := h.cw.SendMessage(API_KEY, ROOM_ID, chatworkString)
	if err != nil {
		return err
	}

	return nil
}

func ConvertSlackPayloadToChatwork(payload SlackPayload) string {
	var builder strings.Builder

	builder.WriteString("[info]\n")

	if payload.Text != "" {
		builder.WriteString("[title]" + payload.Text + "[/title]\n")
	}

	// Handle attachments
	for _, attachment := range payload.Attachments {
		for _, block := range attachment.Blocks {
			if block.Type == "section" {
				results := splitStringToInfo(block.Text.Text)
				for _, info := range results {
					title := addIconsToFields(info.Title)

					if info.Text == "" {
						builder.WriteString(info.Text + "\n")
					} else if info.Title == "" {
						builder.WriteString(info.Text + "\n")
					} else {
						builder.WriteString(title + ": " + info.Text + "\n")
					}
				}
				builder.WriteString("[hr]\n")

			}
		}
	}

	// Handle blocks
	for _, block := range payload.Blocks {
		if block.Type == "section" {
			results := splitStringToInfo(block.Text.Text)
			for _, info := range results {
				title := addIconsToFields(info.Title)
				if info.Text == "" {
					builder.WriteString(info.Text + "\n")
				} else if info.Title == "" {
					builder.WriteString(info.Text + "\n")
				} else {
					builder.WriteString(title + ": " + info.Text + "\n")
				}
			}
			// if the results have any item and the title contains "deployment", add the time
			if len(results) > 0 && strings.Contains(strings.ToLower(results[0].Title), "deployment") {
				builder.WriteString("🕒 Time: " + time.Now().Format("2006-01-02 15:04:05") + "\n")
			}

		}
	}

	// Write footer
	builder.WriteString("\n[hr]\n")
	builder.WriteString("From Slack, sending all our love 💖🤝💬\n")
	builder.WriteString("[/info]\n")

	return builder.String()
}

func splitStringToInfo(text string) []Info {
	text = strings.ReplaceAll(text, `\n`, "\n")
	lines := strings.Split(text, "\n")

	reMarkdownLink := regexp.MustCompile(`(?i)<([^|>]+)\|([^>]+)>`)
	reKeyCandidate := regexp.MustCompile(`^\*?[\w\s\-]+:\s*`)
	reKeyValue := regexp.MustCompile(`^\*?([^:]+?)\*?:\s*(.+)$`)

	var results []Info

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if matches := reMarkdownLink.FindStringSubmatch(line); len(matches) == 3 {
			results = append(results, Info{
				Title: clean(matches[2]),
				Text:  clean(matches[1]),
			})
			continue
		}

		if reKeyCandidate.MatchString(line) {
			if matches := reKeyValue.FindStringSubmatch(line); len(matches) == 3 {
				textValue := matches[2]
				textValue = strings.TrimPrefix(textValue, "* ")
				results = append(results, Info{
					Title: clean(matches[1]),
					Text:  textValue,
				})
				continue
			}
		}

		results = append(results, Info{
			Title: "",
			Text:  "💬 " + line,
		})
	}

	return results
}

func clean(text string) string {
	return strings.Trim(strings.TrimSpace(text), "* ")
}

func addIconsToFields(input string) string {
	replacer := strings.NewReplacer(
		"Application URL", "🔗 URL",
		"Project", "📦 Project",
		"Environment", "🌍 Environment",
		"Deployment Logs", "📄 Deployment Logs",
		"Frequency", "🔄 Frequency",
	)
	return replacer.Replace(input)
}

func (h *HookService) SendToSlack(payload SlackPayload) error {
	webhookURL := utils.GetEnv("SLACK_WEBHOOK_URL", "")
	if webhookURL == "" {
		return fmt.Errorf("slack webhook URL is not configured")
	}

	message := map[string]interface{}{
		"text":   payload.Blocks[0].Text.Text,
		"blocks": payload.Blocks,
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack payload: %w", err)
	}

	req, err := http.NewRequest("POST", webhookURL, strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("failed to create Slack request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Slack request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to send message to Slack: received status %d, response: %s", resp.StatusCode, string(body))
	}

	return nil
}
