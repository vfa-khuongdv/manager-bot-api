package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/repositories"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
)

const chatworkAPIBase = "https://api.chatwork.com/v2"

type IChatworkBotService interface {
	GetAll(paging *utils.Paging) ([]models.BotDetail, int64, error)
	GetBotByID(id uint) (*models.ChatworkBot, error)
	Create(apiToken string, email *string, description string) (*models.BotDetail, error)
	Delete(id uint) error
	GetBotRequests(status string) ([]models.BotRequestItem, error)
	AcceptBotRequest(compositeID string) error
	DeleteBotRequest(compositeID string) error
}

type ChatworkBotService struct {
	repo       repositories.IChatworkBotRepository
	httpClient *http.Client
}

func NewChatworkBotService(repo repositories.IChatworkBotRepository) *ChatworkBotService {
	return &ChatworkBotService{
		repo:       repo,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// GetAll returns all bots enriched with live Chatwork profile data.
func (s *ChatworkBotService) GetAll(paging *utils.Paging) ([]models.BotDetail, int64, error) {
	bots, total, err := s.repo.GetAll(paging)
	if err != nil {
		return nil, 0, err
	}

	details := make([]models.BotDetail, 0, len(bots))
	for i := range bots {
		detail := s.enrichBot(&bots[i])
		details = append(details, detail)
	}
	return details, total, nil
}

// GetBotByID fetches a bot record from the database by its ID.
func (s *ChatworkBotService) GetBotByID(id uint) (*models.ChatworkBot, error) {
	return s.repo.GetByID(id)
}

// Create registers a new bot, validates the token against Chatwork, then persists it.
func (s *ChatworkBotService) Create(apiToken string, email *string, description string) (*models.BotDetail, error) {
	// Validate token is working before saving
	profile := s.fetchMe(apiToken)
	if profile == nil {
		return nil, fmt.Errorf("invalid or unauthorized Chatwork API token")
	}

	bot := &models.ChatworkBot{
		APIToken:    apiToken,
		Email:       email,
		Description: description,
	}
	created, err := s.repo.Create(bot)
	if err != nil {
		return nil, err
	}

	rooms := s.fetchRooms(created.APIToken)
	detail := s.buildBotDetail(created, profile, len(rooms))
	return &detail, nil
}

// Delete removes a bot from the system by its DB ID.
func (s *ChatworkBotService) Delete(id uint) error {
	if _, err := s.repo.GetByID(id); err != nil {
		return fmt.Errorf("bot not found")
	}
	return s.repo.Delete(id)
}

// GetBotRequests fetches incoming requests from all bots via Chatwork API.
// status filter: "" = all (pending only on CW side), or "pending" explicitly.
func (s *ChatworkBotService) GetBotRequests(status string) ([]models.BotRequestItem, error) {
	// Get all bots (no paging limit — use a large limit)
	allPaging := &utils.Paging{Page: 1, Limit: 1000}
	bots, _, err := s.repo.GetAll(allPaging)
	if err != nil {
		return nil, err
	}

	var items []models.BotRequestItem
	for i := range bots {
		bot := &bots[i]
		requests, err := s.fetchIncomingRequests(bot.APIToken)
		if err != nil {
			logger.Errorf("[BotRequests] bot_id=%d fetchIncomingRequests failed: %v", bot.ID, err)
			return nil, fmt.Errorf("bot_id=%d: %w", bot.ID, err)
		}

		profile := s.fetchMe(bot.APIToken)
		botDetail := s.buildBotDetail(bot, profile, 0)

		for _, req := range requests {
			// Chatwork incoming_requests are always "pending" status
			// (accepted/rejected ones disappear from the list)
			reqStatus := "pending"
			if status != "" && status != reqStatus {
				continue
			}
			items = append(items, models.BotRequestItem{
				ID:        fmt.Sprintf("%d_%d", bot.ID, req.RequestID),
				BotID:     bot.ID,
				BotInfo:   &botDetail,
				Status:    reqStatus,
				CreatedAt: time.Now().UTC().Format(time.RFC3339), // CW API does not return createdAt
			})
		}
	}
	return items, nil
}

// AcceptBotRequest accepts a Chatwork incoming request.
// compositeID format: "{dbBotID}_{cwRequestID}"
func (s *ChatworkBotService) AcceptBotRequest(compositeID string) error {
	botID, cwReqID, err := parseCompositeID(compositeID)
	if err != nil {
		return err
	}
	bot, err := s.repo.GetByID(botID)
	if err != nil {
		return fmt.Errorf("bot not found: %w", err)
	}

	url := fmt.Sprintf("%s/incoming_requests/%d", chatworkAPIBase, cwReqID)
	req, _ := http.NewRequest(http.MethodPut, url, nil)
	req.Header.Set("X-ChatWorkToken", bot.APIToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("chatwork API error: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("chatwork accept failed (status %d): %s", resp.StatusCode, string(body))
	}
	return nil
}

// DeleteBotRequest rejects/deletes a Chatwork incoming request.
// compositeID format: "{dbBotID}_{cwRequestID}"
func (s *ChatworkBotService) DeleteBotRequest(compositeID string) error {
	botID, cwReqID, err := parseCompositeID(compositeID)
	if err != nil {
		return err
	}
	bot, err := s.repo.GetByID(botID)
	if err != nil {
		return fmt.Errorf("bot not found: %w", err)
	}

	url := fmt.Sprintf("%s/incoming_requests/%d", chatworkAPIBase, cwReqID)
	req, _ := http.NewRequest(http.MethodDelete, url, nil)
	req.Header.Set("X-ChatWorkToken", bot.APIToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("chatwork API error: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("chatwork delete failed (status %d): %s", resp.StatusCode, string(body))
	}
	return nil
}

// ── private helpers ───────────────────────────────────────────────────────────

func (s *ChatworkBotService) enrichBot(bot *models.ChatworkBot) models.BotDetail {
	profile := s.fetchMe(bot.APIToken)
	rooms := s.fetchRooms(bot.APIToken)
	return s.buildBotDetail(bot, profile, len(rooms))
}

func (s *ChatworkBotService) buildBotDetail(bot *models.ChatworkBot, profile *models.ChatworkMeResponse, roomsCount int) models.BotDetail {
	detail := models.BotDetail{
		ID:          bot.ID,
		Email:       bot.Email,
		Description: bot.Description,
		RoomsCount:  roomsCount,
	}
	if profile != nil {
		detail.AccountID = profile.AccountID
		detail.ChatworkID = profile.ChatworkID
		detail.Name = profile.Name
		detail.AvatarURL = profile.AvatarImageURL
	}
	return detail
}

func (s *ChatworkBotService) fetchMe(apiToken string) *models.ChatworkMeResponse {
	req, _ := http.NewRequest(http.MethodGet, chatworkAPIBase+"/me", nil)
	req.Header.Set("X-ChatWorkToken", apiToken)

	resp, err := s.httpClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil
	}
	defer resp.Body.Close()

	var me models.ChatworkMeResponse
	if err := json.NewDecoder(resp.Body).Decode(&me); err != nil {
		return nil
	}
	return &me
}

func (s *ChatworkBotService) fetchRooms(apiToken string) []models.ChatworkRoom {
	req, _ := http.NewRequest(http.MethodGet, chatworkAPIBase+"/rooms", nil)
	req.Header.Set("X-ChatWorkToken", apiToken)

	resp, err := s.httpClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil
	}
	defer resp.Body.Close()

	var rooms []models.ChatworkRoom
	json.NewDecoder(resp.Body).Decode(&rooms) //nolint:errcheck
	return rooms
}

func (s *ChatworkBotService) fetchIncomingRequests(apiToken string) ([]models.ChatworkIncomingRequest, error) {
	req, _ := http.NewRequest(http.MethodGet, chatworkAPIBase+"/incoming_requests", nil)
	req.Header.Set("X-ChatWorkToken", apiToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Chatwork returns 204 No Content when there are no pending requests
	if resp.StatusCode == http.StatusNoContent {
		return []models.ChatworkIncomingRequest{}, nil
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("chatwork API %d: %s", resp.StatusCode, string(body))
	}

	var reqs []models.ChatworkIncomingRequest
	if err := json.NewDecoder(resp.Body).Decode(&reqs); err != nil {
		return nil, err
	}
	return reqs, nil
}

// parseCompositeID splits "{botID}_{cwRequestID}" into typed values.
func parseCompositeID(id string) (uint, int, error) {
	var botID uint
	var cwReqID int
	_, err := fmt.Sscanf(strings.Replace(id, "_", " ", 1), "%d %d", &botID, &cwReqID)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid request ID format (expected {botID}_{cwRequestID}): %w", err)
	}
	return botID, cwReqID, nil
}
