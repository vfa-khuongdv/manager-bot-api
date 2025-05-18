package services

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/vfa-khuongdv/golang-cms/internal/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
)

type IHookService interface {
	ChatworkHook(payload DiscordPayload) error
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
			builder.WriteString("\n[hr]\n" + embed.Footer.Text + "\n")
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
