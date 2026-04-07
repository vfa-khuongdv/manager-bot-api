package main

import (
	"fmt"
	"os"

	"github.com/vfa-khuongdv/golang-cms/internal/configs"
	"github.com/vfa-khuongdv/golang-cms/internal/services"
	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
)

func main() {
	configs.LoadEnv()
	logger.Init()

	roomID := os.Getenv("CVE_CHATWORK_ROOM_ID")
	apiKey := os.Getenv("CVE_CHATWORK_API_KEY")
	nvdAPIKey := os.Getenv("NVD_API_KEY")

	if roomID == "" || apiKey == "" {
		fmt.Println("Please set CVE_CHATWORK_ROOM_ID and CVE_CHATWORK_API_KEY in .env")
		os.Exit(1)
	}

	fmt.Printf("Testing CVE Crawler...\n")
	fmt.Printf("Room ID: %s\n", roomID)
	fmt.Printf("NVD API Key: %s\n", nvdAPIKey)

	cveService := services.NewCveCrawlerService(roomID, apiKey, nvdAPIKey)
	cveService.CrawlAndNotify()

	fmt.Println("Done!")
}
