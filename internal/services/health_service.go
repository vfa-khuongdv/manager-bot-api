package services

import (
	"fmt"
	"math"
	"net/http"
	"runtime"
	"time"

	"github.com/vfa-khuongdv/golang-cms/internal/configs"
	"github.com/vfa-khuongdv/golang-cms/internal/models"
	"golang.org/x/sys/unix"
)

type ChatworkHealth struct {
	Status      string `json:"status"`
	Latency     int64  `json:"latency"`
	APIVersion  string `json:"apiVersion"`
	LastChecked string `json:"lastChecked"`
}

type ServerHealth struct {
	Status string       `json:"status"`
	Uptime int64        `json:"uptime"`
	CPU    CPUHealth    `json:"cpu"`
	Memory MemoryHealth `json:"memory"`
}

type CPUHealth struct {
	Usage float64 `json:"usage"`
}

type MemoryHealth struct {
	Used  int64 `json:"used"`
	Total int64 `json:"total"`
}

type DatabaseHealth struct {
	Status  string     `json:"status"`
	Latency int64      `json:"latency"`
	Pool    PoolHealth `json:"pool"`
}

type PoolHealth struct {
	Active int `json:"active"`
	Idle   int `json:"idle"`
	Total  int `json:"total"`
}

type OverallHealthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Checks    map[string]string `json:"checks"`
}

type HealthChatworkService struct {
	BaseURL string
}

func NewHealthChatworkService() *HealthChatworkService {
	return &HealthChatworkService{
		BaseURL: "https://api.chatwork.com/v2",
	}
}

func (s *HealthChatworkService) CheckHealth() (ChatworkHealth, error) {
	start := time.Now()

	db := configs.DB
	if db == nil {
		return ChatworkHealth{
			Status:      "down",
			Latency:     0,
			APIVersion:  "v2",
			LastChecked: time.Now().Format(time.RFC3339),
		}, fmt.Errorf("database not initialized")
	}

	var bot models.ChatworkBot
	if err := db.Table("chatwork_bots").First(&bot).Error; err != nil {
		return ChatworkHealth{
			Status:      "down",
			Latency:     0,
			APIVersion:  "v2",
			LastChecked: time.Now().Format(time.RFC3339),
		}, fmt.Errorf("no chatwork bot found: %w", err)
	}

	apiKey := bot.APIToken
	if apiKey == "" {
		return ChatworkHealth{
			Status:      "down",
			Latency:     0,
			APIVersion:  "v2",
			LastChecked: time.Now().Format(time.RFC3339),
		}, fmt.Errorf("chatwork bot api_token is empty")
	}

	req, err := http.NewRequest("GET", s.BaseURL+"/me", nil)
	if err != nil {
		return ChatworkHealth{
			Status:      "down",
			APIVersion:  "v2",
			LastChecked: time.Now().Format(time.RFC3339),
		}, err
	}

	req.Header.Set("X-ChatWorkToken", apiKey)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return ChatworkHealth{
			Status:      "down",
			Latency:     latency,
			APIVersion:  "v2",
			LastChecked: time.Now().Format(time.RFC3339),
		}, err
	}
	defer resp.Body.Close()

	status := "operational"
	if latency >= 500 {
		status = "degraded"
	}

	return ChatworkHealth{
		Status:      status,
		Latency:     latency,
		APIVersion:  "v2",
		LastChecked: time.Now().Format(time.RFC3339),
	}, nil
}

var serverStartTime int64 = 0

func init() {
	serverStartTime = time.Now().Unix()
}

func GetServerHealth() ServerHealth {
	var r unix.Rusage
	unix.Getrusage(unix.RUSAGE_SELF, &r)

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	usedMem := m.Alloc
	totalMem := uint64(r.Maxrss * 1024)
	if totalMem == 0 {
		totalMem = 1
	}
	memUsage := float64(usedMem) / float64(totalMem) * 100

	utime := float64(r.Utime.Sec) + float64(r.Utime.Usec)/1000000
	stime := float64(r.Stime.Sec) + float64(r.Stime.Usec)/1000000
	totalCpuTime := utime + stime

	uptimeSeconds := time.Now().Unix() - serverStartTime
	if uptimeSeconds <= 0 {
		uptimeSeconds = 1
	}

	cpuUsage := (totalCpuTime / float64(uptimeSeconds)) * 100
	if cpuUsage > 100 {
		cpuUsage = 100
	}

	var status string
	if cpuUsage > 95 || memUsage > 95 {
		status = "unhealthy"
	} else if cpuUsage > 80 || memUsage > 80 {
		status = "degraded"
	} else {
		status = "healthy"
	}

	return ServerHealth{
		Status: status,
		Uptime: uptimeSeconds,
		CPU: CPUHealth{
			Usage: math.Round(cpuUsage*100) / 100,
		},
		Memory: MemoryHealth{
			Used:  int64(usedMem / (1024 * 1024)),
			Total: int64(totalMem / (1024 * 1024)),
		},
	}
}

func GetDatabaseHealth() (DatabaseHealth, error) {
	db := configs.DB
	if db == nil {
		return DatabaseHealth{
			Status:  "disconnected",
			Latency: 0,
			Pool: PoolHealth{
				Active: 0,
				Idle:   0,
				Total:  0,
			},
		}, fmt.Errorf("database not initialized")
	}

	start := time.Now()
	sqlDB, err := db.DB()
	if err != nil {
		return DatabaseHealth{
			Status:  "disconnected",
			Latency: 0,
			Pool: PoolHealth{
				Active: 0,
				Idle:   0,
				Total:  0,
			},
		}, err
	}

	if err := sqlDB.Ping(); err != nil {
		return DatabaseHealth{
			Status:  "disconnected",
			Latency: time.Since(start).Milliseconds(),
			Pool: PoolHealth{
				Active: 0,
				Idle:   0,
				Total:  0,
			},
		}, err
	}

	latency := time.Since(start).Milliseconds()

	stats := sqlDB.Stats()
	poolMetrics := PoolHealth{
		Active: stats.InUse,
		Idle:   stats.Idle,
		Total:  stats.MaxOpenConnections,
	}

	poolUsage := float64(poolMetrics.Active) / float64(poolMetrics.Total) * 100

	status := "connected"
	if latency >= 1000 || poolUsage > 80 {
		status = "degraded"
	}

	return DatabaseHealth{
		Status:  status,
		Latency: latency,
		Pool:    poolMetrics,
	}, nil
}

func GetOverallHealth() OverallHealthResponse {
	chatworkStatus := "up"
	serverStatus := "up"
	databaseStatus := "up"

	chatworkHealth, err := NewHealthChatworkService().CheckHealth()
	if err != nil || chatworkHealth.Status == "down" {
		chatworkStatus = "down"
	} else if chatworkHealth.Status == "degraded" {
		chatworkStatus = "degraded"
	}

	serverHealth := GetServerHealth()
	if serverHealth.Status == "unhealthy" {
		serverStatus = "down"
	} else if serverHealth.Status == "degraded" {
		serverStatus = "degraded"
	}

	databaseHealth, err := GetDatabaseHealth()
	if err != nil || databaseHealth.Status == "disconnected" {
		databaseStatus = "down"
	} else if databaseHealth.Status == "degraded" {
		databaseStatus = "degraded"
	}

	checks := map[string]string{
		"chatwork": chatworkStatus,
		"server":   serverStatus,
		"database": databaseStatus,
	}

	statuses := []string{chatworkStatus, serverStatus, databaseStatus}
	overallStatus := "healthy"

	hasDown := false
	hasDegraded := false
	for _, s := range statuses {
		if s == "down" {
			hasDown = true
		}
		if s == "degraded" {
			hasDegraded = true
		}
	}

	if hasDown {
		overallStatus = "unhealthy"
	} else if hasDegraded {
		overallStatus = "degraded"
	}

	return OverallHealthResponse{
		Status:    overallStatus,
		Timestamp: time.Now().Format(time.RFC3339),
		Checks:    checks,
	}
}
