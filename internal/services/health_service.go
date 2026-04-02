package services

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/vfa-khuongdv/golang-cms/internal/configs"
)

var serverStartTime int64 = 0

func init() {
	serverStartTime = time.Now().Unix()
}

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

	apiKey := os.Getenv("CHATWORK_API_TOKEN")
	if apiKey == "" {
		return ChatworkHealth{
			Status:      "down",
			Latency:     0,
			APIVersion:  "v2",
			LastChecked: time.Now().Format(time.RFC3339),
		}, fmt.Errorf("CHATWORK_API_TOKEN not configured")
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

func GetServerHealth() ServerHealth {
	cpuPercents, _ := cpu.Percent(0, false)
	cpuPercent := 0.0
	if len(cpuPercents) > 0 {
		cpuPercent = cpuPercents[0]
	}

	v, _ := mem.VirtualMemory()

	uptime := time.Now().Unix() - serverStartTime
	if uptime <= 0 {
		uptime = 1
	}

	var status string
	if cpuPercent > 95 || v.UsedPercent > 95 {
		status = "unhealthy"
	} else if cpuPercent > 80 || v.UsedPercent > 80 {
		status = "degraded"
	} else {
		status = "healthy"
	}

	return ServerHealth{
		Status: status,
		Uptime: uptime,
		CPU: CPUHealth{
			Usage: cpuPercent,
		},
		Memory: MemoryHealth{
			Used:  int64(v.Used) / (1024 * 1024),
			Total: int64(v.Total) / (1024 * 1024),
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
