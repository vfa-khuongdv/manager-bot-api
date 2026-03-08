package repositories

import (
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
	"gorm.io/gorm"
)

type IChatworkBotRepository interface {
	GetAll(paging *utils.Paging) ([]models.ChatworkBot, int64, error)
	GetByID(id uint) (*models.ChatworkBot, error)
	Create(bot *models.ChatworkBot) (*models.ChatworkBot, error)
	Delete(id uint) error
}

type ChatworkBotRepository struct {
	db *gorm.DB
}

func NewChatworkBotRepository(db *gorm.DB) *ChatworkBotRepository {
	return &ChatworkBotRepository{db: db}
}

func (r *ChatworkBotRepository) GetAll(paging *utils.Paging) ([]models.ChatworkBot, int64, error) {
	var bots []models.ChatworkBot
	q := r.db.Model(&models.ChatworkBot{})

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (paging.Page - 1) * paging.Limit
	if err := q.Offset(offset).Limit(paging.Limit).Find(&bots).Error; err != nil {
		return nil, 0, err
	}
	return bots, total, nil
}

func (r *ChatworkBotRepository) GetByID(id uint) (*models.ChatworkBot, error) {
	var bot models.ChatworkBot
	if err := r.db.First(&bot, id).Error; err != nil {
		return nil, err
	}
	return &bot, nil
}

func (r *ChatworkBotRepository) Create(bot *models.ChatworkBot) (*models.ChatworkBot, error) {
	if err := r.db.Create(bot).Error; err != nil {
		return nil, err
	}
	return bot, nil
}

func (r *ChatworkBotRepository) Delete(id uint) error {
	return r.db.Delete(&models.ChatworkBot{}, id).Error
}
