package models

import (
	"time"

	"gorm.io/gorm"
)

// ChatworkBot is persisted in DB — stores only what we own.
// Profile data (name, accountId, avatarUrl, etc.) is fetched live from Chatwork API.
type ChatworkBot struct {
	ID          uint           `json:"id"`
	APIToken    string         `json:"-" gorm:"column:api_token;type:varchar(255);not null"`
	Email       *string        `json:"email,omitempty" gorm:"type:varchar(255)"`
	Description string         `json:"description" gorm:"type:text"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `json:"deletedAt,omitempty"`
}

func (ChatworkBot) TableName() string {
	return "chatwork_bots"
}

// ChatworkMeResponse maps the Chatwork GET /v2/me API response.
type ChatworkMeResponse struct {
	AccountID        int    `json:"account_id"`
	RoomID           int    `json:"room_id"`
	Name             string `json:"name"`
	ChatworkID       string `json:"chatwork_id"`
	OrganizationID   int    `json:"organization_id"`
	OrganizationName string `json:"organization_name"`
	Department       string `json:"department"`
	Title            string `json:"title"`
	URL              string `json:"url"`
	Introduction     string `json:"introduction"`
	Mail             string `json:"mail"`
	AvatarImageURL   string `json:"avatar_image_url"`
	LoginMail        string `json:"login_mail"`
}

// ChatworkRoom maps a minimal entry from GET /v2/rooms.
type ChatworkRoom struct {
	RoomID     int `json:"room_id"`
	MessageNum int `json:"message_num"`
}

// ChatworkIncomingRequest maps an entry from GET /v2/incoming_requests.
type ChatworkIncomingRequest struct {
	RequestID        int    `json:"request_id"`
	AccountID        int    `json:"account_id"`
	Message          string `json:"message"`
	Name             string `json:"name"`
	ChatworkID       string `json:"chatwork_id"`
	OrganizationID   int    `json:"organization_id"`
	OrganizationName string `json:"organization_name"`
	Department       string `json:"department"`
	AvatarImageURL   string `json:"avatar_image_url"`
}

// BotDetail is the enriched API response combining DB + live Chatwork profile data.
type BotDetail struct {
	ID          uint    `json:"id"`
	AccountID   int     `json:"accountId"`
	ChatworkID  string  `json:"chatworkId"`
	Name        string  `json:"name"`
	Email       *string `json:"email,omitempty"`
	AvatarURL   string  `json:"avatarUrl"`
	Description string  `json:"description"`
	RoomsCount  int     `json:"roomsCount"`
}

// SenderInfo holds the Chatwork profile of the person who sent the friend request.
type SenderInfo struct {
	AccountID        int    `json:"accountId"`
	Name             string `json:"name"`
	ChatworkID       string `json:"chatworkId"`
	OrganizationName string `json:"organizationName"`
	Department       string `json:"department"`
	AvatarImageURL   string `json:"avatarImageUrl"`
}

// BotRequestItem is the response item for GET /bot-requests.
// The ID is encoded as "{dbBotID}_{cwRequestID}" so accept/delete can
// route to the correct bot API token without a DB lookup table.
type BotRequestItem struct {
	ID         string      `json:"id"`
	BotID      uint        `json:"botId"`
	BotInfo    *BotDetail  `json:"botInfo"`
	SenderInfo *SenderInfo `json:"senderInfo"`
	Message    string      `json:"message"`
	Status     string      `json:"status"`
	CreatedAt  string      `json:"createdAt"`
}
