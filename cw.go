package main

import (
	"fmt"
	"regexp"
	"strings"
)

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

func clean(text string) string {
	return strings.Trim(text, "* ")
}

func splitStringToInfo(text string) {
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

		// results = append(results, Info{
		// 	Title: line,
		// 	Text:  "",
		// })
	}

	for _, info := range results {
		fmt.Printf("Title: %s\nText: %s\n\n", info.Title, info.Text)
	}
}

func main() {
	text1 := `New version successfully deployed for vfa-khuongdv/manager-bot-api:main-eosg0gsg0sso8coo0g4kww40\nApplication URL: https://manager-bot-api.vfa-spl.org\n\n*Project:* manager-bot\n*Environment:* production\n*<https://coolify.vfa-spl.org/project/tcgkcsg8804soo8gcocw08g0/environment/usgwwkogo484kowcokcc0wgw/application/eosg0gsg0sso8coo0g4kww40/deployment/qw40owwsok4gg8o08g8o004g|Deployment Logs>*`

	text2 := `Database backup for Database (db:forum) was successful.\n\n*Frequency:* 0 0 * * *`

	fmt.Println("=== Text 1 ===")
	splitStringToInfo(text1)

	fmt.Println("=== Text 2 ===")
	splitStringToInfo(text2)
}
