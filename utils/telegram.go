// utils/telegram.go
package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type TelegramChatResponse struct {
	OK     bool `json:"ok"`
	Result struct {
		ID int64 `json:"id"`
	} `json:"result"`
	Description string `json:"description"`
}

// ExtractUsernameFromLink parses "https://t.me/mygroup" → "@mygroup"
func ExtractUsernameFromLink(link string) (string, error) {
	link = strings.TrimSpace(link)

	// Already a username like @mygroup
	// check wether have @ or not
	if strings.HasPrefix(link, "@") {
		return link, nil
	}

	// Handle https://t.me/mygroup or t.me/mygroup
	link = strings.TrimPrefix(link, "https://")
	link = strings.TrimPrefix(link, "http://")
	link = strings.TrimPrefix(link, "t.me/")

	if link == "" {
		return "", errors.New("invalid telegram group link")
	}

	// Private invite links (t.me/+xxxxx) cannot be resolved via getChat
	if strings.HasPrefix(link, "+") {
		return "", errors.New("private invite links are not supported, please use a public group link")
	}

	return "@" + link, nil
}

// ResolveTelegramChatID calls Telegram API to get Chat ID from a group link
func ResolveTelegramChatID(botToken, groupLink string) (int64, error) {
	username, err := ExtractUsernameFromLink(groupLink)
	if err != nil {
		return 0, err
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/getChat?chat_id=%s", botToken, username)

	resp, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to contact telegram API: %w", err)
	}
	defer resp.Body.Close()

	var result TelegramChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to decode telegram response: %w", err)
	}

	if !result.OK {
		return 0, fmt.Errorf("telegram API error: %s", result.Description)
	}

	return result.Result.ID, nil
}
