package services

import (
	"fmt"
	"sync"

	"github.com/robfig/cron/v3"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
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
		// Mask token for logging
		maskedToken := ""
		if len(s.ChatworkToken) > 4 {
			maskedToken = s.ChatworkToken[:4] + "..."
		} else if len(s.ChatworkToken) > 0 {
			maskedToken = "..." // Mask if token is short but not empty
		} else {
			maskedToken = "[EMPTY]" // Indicate if token is empty
		}

		logger.Infof("[Reminder #%d] Attempting to send message. RoomID: '%s', Token (masked): '%s', Message: '%s'", s.ID, s.ChatworkRoomID, maskedToken, s.Message)

		err := cs.cw.SendMessage(s.ChatworkToken, s.ChatworkRoomID, s.Message)
		if err != nil {
			// Log detailed error. The error 'err' from cs.cw.SendMessage should ideally include details from Chatwork API.
			logger.Errorf("[Reminder #%d] Error sending message to RoomID '%s'. Token (masked): '%s', Message: '%s'. Error: %v", s.ID, s.ChatworkRoomID, maskedToken, s.Message, err)
		} else {
			logger.Infof("[Reminder #%d] Successfully sent message to room %s", s.ID, s.ChatworkRoomID)
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
	fmt.Println("Cron service started")
}

// Stop stops the cron scheduler
func (cs *CronService) Stop() {
	cs.c.Stop()
	fmt.Println("Cron service stopped")
}

// SyncAll synchronizes all active reminder schedules from the database
func (cs *CronService) SyncAll() {

	logger.Infof("Syncing all active reminder schedules from the database")

	cs.lock.Lock()
	defer cs.lock.Unlock()

	var schedules []models.ReminderSchedule
	result := cs.db.Where("active = ?", true).Find(&schedules)
	if result.Error != nil {
		fmt.Printf("Error loading reminder schedules: %v\n", result.Error)
		return
	}

	for _, s := range schedules {
		sCopy := s
		cs.Register(&sCopy)
	}

	// show all active schedules
	fmt.Printf("Active schedules: %v\n", cs.entries)
	for id, entryID := range cs.entries {
		fmt.Printf("Schedule ID: %d, Entry ID: %d\n", id, entryID)
	}
}
