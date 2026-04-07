package services

import (
	"sync"

	"github.com/robfig/cron/v3"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
	"gorm.io/gorm"
)

type CronService struct {
	c       *cron.Cron
	entries map[uint]cron.EntryID
	db      *gorm.DB
	lock    sync.Mutex
	cw      *ChatworkService
}

type ICronService interface {
	LoadFromDB()
	Register(s *models.ReminderSchedule)
	Remove(scheduleID uint)
	Start()
	Stop()
	SyncAll()
	RegisterCVECrawler()
}

func NewCronService(db *gorm.DB) *CronService {
	return &CronService{
		c:       cron.New(cron.WithSeconds()),
		entries: make(map[uint]cron.EntryID),
		db:      db,
		cw:      NewChatworkService(),
	}
}

func (cs *CronService) LoadFromDB() {
	var schedules []models.ReminderSchedule
	result := cs.db.Where("active = ?", true).Find(&schedules)
	if result.Error != nil {
		logger.Errorf("Error loading reminder schedules: %v", result.Error) // Changed to logger
		return
	}

	logger.Infof("Found %d active reminder schedules", len(schedules)) // Changed to logger
	for _, s := range schedules {
		sCopy := s
		cs.Register(&sCopy)
		// Removed redundant log, Register will log its own status
	}
}

func (cs *CronService) Register(s *models.ReminderSchedule) {
	cs.lock.Lock()
	defer cs.lock.Unlock()

	// Remove existing schedule if it exists
	cs.Remove(s.ID)

	entryID, err := cs.c.AddFunc(s.CronExpression, func() {
		// Resolve the token to use: prefer bot's token if botId is set
		token := ""
		if s.ChatworkToken != nil {
			token = *s.ChatworkToken
		}
		if s.BotID != nil {
			var bot models.ChatworkBot
			if err := cs.db.First(&bot, *s.BotID).Error; err != nil {
				logger.Errorf("[Reminder #%d] Failed to fetch bot token for BotID %d: %v", s.ID, *s.BotID, err)
			} else {
				token = bot.APIToken
			}
		}

		// Mask token for logging
		maskedToken := ""
		if len(token) > 4 {
			maskedToken = token[:4] + "..."
		} else if len(token) > 0 {
			maskedToken = "..." // Mask if token is short but not empty
		} else {
			maskedToken = "[EMPTY]" // Indicate if token is empty
		}

		logger.Infof("[Reminder #%d] Attempting to send message. RoomID: '%s', Token (masked): '%s', Message: '%s'", s.ID, s.ChatworkRoomID, maskedToken, s.Message)

		err := cs.cw.SendMessage(token, s.ChatworkRoomID, s.Message)

		logEntry := models.ScheduleLog{
			ProjectID:  s.ProjectID,
			ScheduleID: s.ID,
			Status:     "success",
		}

		if err != nil {
			// Log detailed error. The error 'err' from cs.cw.SendMessage should ideally include details from Chatwork API.
			logger.Errorf("[Reminder #%d] Error sending message to RoomID '%s'. Token (masked): '%s', Message: '%s'. Error: %v", s.ID, s.ChatworkRoomID, maskedToken, s.Message, err)
			logEntry.Status = "error"
			logEntry.ErrorMessage = err.Error()
		} else {
			logger.Infof("[Reminder #%d] Successfully sent message to room %s", s.ID, s.ChatworkRoomID)
		}

		if dbErr := cs.db.Create(&logEntry).Error; dbErr != nil {
			logger.Errorf("[Reminder #%d] Error recording schedule log: %v", s.ID, dbErr)
		}
	})

	if err != nil {
		logger.Errorf("Error registering cron job for reminder #%d with expression '%s': %v", s.ID, s.CronExpression, err)
	} else {
		cs.entries[s.ID] = entryID
		logger.Infof("Successfully registered cron job for reminder #%d with ID %d and expression: %s", s.ID, entryID, s.CronExpression)
	}
}

func (cs *CronService) Remove(scheduleID uint) {
	if id, ok := cs.entries[scheduleID]; ok {
		cs.c.Remove(id)
		delete(cs.entries, scheduleID)
	}
}

// Start begins running the cron scheduler
func (cs *CronService) Start() {
	cs.c.Start()
	logger.Info("Cron service started")
}

// Stop stops the cron scheduler
func (cs *CronService) Stop() {
	cs.c.Stop()
	logger.Info("Cron service stopped")
}

// 1. Get all active schedules of each project
// 2. Unregister all existing schedules in the cron service
// 3. Register all active schedules again (with updated cron expressions or messages)
func (cs *CronService) SyncAll() {
	logger.Info("Synchronizing all reminder schedules")

	var projects []models.Project
	if err := cs.db.Preload("ReminderSchedules", "active = ?", true).Find(&projects).Error; err != nil {
		logger.Errorf("Error fetching projects for synchronization: %v", err)
		return
	}

	for scheduleID := range cs.entries {
		logger.Infof("Removing existing cron job for schedule ID %d", scheduleID)
		cs.Remove(scheduleID)
	}

	for _, project := range projects {
		for _, schedule := range project.ReminderSchedules {
			sCopy := schedule
			logger.Infof("Registering cron job for schedule ID %d (Project: %s, Expression: %s)", schedule.ID, project.Name, schedule.CronExpression)
			cs.Register(&sCopy)
		}
	}
}

const cveJobEntryID = 99999

func (cs *CronService) RegisterCVECrawler() {
	roomID := utils.GetEnv("CVE_CHATWORK_ROOM_ID", "")
	apiKey := utils.GetEnv("CVE_CHATWORK_API_KEY", "")
	nvdAPIKey := utils.GetEnv("NVD_API_KEY", "")

	if roomID == "" || apiKey == "" {
		logger.Warn("[CVE] CVE_CHATWORK_ROOM_ID or CVE_CHATWORK_API_KEY not configured, skipping CVE crawler job")
		return
	}

	cveService := NewCveCrawlerService(roomID, apiKey, nvdAPIKey)

	_, err := cs.c.AddFunc("0 0 0 * * *", func() {
		logger.Info("[CVE] Starting daily CVE crawl job")
		cveService.CrawlAndNotify()
	})

	if err != nil {
		logger.Errorf("[CVE] Failed to register CVE crawler job: %v", err)
		return
	}

	cs.entries[cveJobEntryID] = cron.EntryID(cveJobEntryID)
	logger.Info("[CVE] CVE crawler job registered successfully (runs daily at 00:00 UTC / 07:00 GMT+7)")
}
