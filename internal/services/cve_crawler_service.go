package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/vfa-khuongdv/golang-cms/internal/constants"
	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
)

type NVDTime time.Time

func (nt *NVDTime) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	if s == "" {
		*nt = NVDTime(time.Time{})
		return nil
	}

	formats := []string{
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05.000-07:00",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
	}

	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			*nt = NVDTime(t)
			return nil
		}
	}

	return fmt.Errorf("unable to parse time: %s", s)
}

type NVDResponse struct {
	ResultsPerPage  int                `json:"resultsPerPage"`
	StartIndex      int                `json:"startIndex"`
	TotalResults    int                `json:"totalResults"`
	Vulnerabilities []NVDVulnerability `json:"vulnerabilities"`
}

type NVDVulnerability struct {
	CVE NVDCVE `json:"cve"`
}

type NVDCVE struct {
	ID           string          `json:"id"`
	Published    NVDTime         `json:"published"`
	LastModified NVDTime         `json:"lastModified"`
	Description  []NVDescription `json:"descriptions"`
	Metrics      NVMetrics       `json:"metrics"`
}

type NVDescription struct {
	Lang  string `json:"lang"`
	Value string `json:"value"`
}

type NVMetrics struct {
	CvssMetricV31 []CVSSMetric   `json:"cvssMetricV31,omitempty"`
	CvssMetricV30 []CVSSMetric   `json:"cvssMetricV30,omitempty"`
	CvssMetricV2  []CVSSMetricV2 `json:"cvssMetricV2,omitempty"`
}

type CVSSMetric struct {
	CVSSData CVSSData `json:"cvssData"`
}

type CVSSData struct {
	BaseScore    float64 `json:"baseScore"`
	BaseSeverity string  `json:"baseSeverity"`
}

type CVSSMetricV2 struct {
	CVSSData CVSSDataV2 `json:"cvssData"`
}

type CVSSDataV2 struct {
	BaseScore    float64 `json:"baseScore"`
	BaseSeverity string  `json:"baseSeverity"`
}

type ICveCrawlerService interface {
	CrawlAndNotify()
}

type CveCrawlerService struct {
	cw        *ChatworkService
	roomID    string
	apiKey    string
	nvdAPIKey string
	languages []string
}

func NewCveCrawlerService(roomID, apiKey, nvdAPIKey string) *CveCrawlerService {
	return &CveCrawlerService{
		cw:        NewChatworkService(),
		roomID:    roomID,
		apiKey:    apiKey,
		nvdAPIKey: nvdAPIKey,
		languages: constants.CVELanguages,
	}
}

func (s *CveCrawlerService) CrawlAndNotify() {
	logger.Info("Starting CVE daily crawl...")

	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	dateFormat := "2006-01-02T15:04:05.000Z"

	pubStartDate := yesterday.Format(dateFormat)
	pubEndDate := now.Format(dateFormat)

	allItems, err := s.fetchAllCVEs(pubStartDate, pubEndDate)
	if err != nil {
		logger.Errorf("[CVE] Error fetching CVEs: %v", err)
		s.cw.SendMessage(s.apiKey, s.roomID, fmt.Sprintf("[info][title]🚨 CVE ERROR[/title]\nLỗi khi fetch CVE: %v\n[/info]", err))
		return
	}

	if len(allItems) == 0 {
		logger.Info("No new CVEs found in the last 24 hours")
		return
	}

	message := s.formatMessageByScore(allItems, now)
	if message == "" {
		logger.Info("No CVEs to report")
		return
	}

	if err := s.cw.SendMessage(s.apiKey, s.roomID, message); err != nil {
		logger.Errorf("[CVE] Failed to send message to Chatwork: %v", err)
	} else {
		logger.Infof("[CVE] Successfully sent CVE report to Chatwork (%d CVEs)", len(allItems))
	}
}

func (s *CveCrawlerService) fetchAllCVEs(pubStartDate, pubEndDate string) ([]CVEItem, error) {
	baseURL := "https://services.nvd.nist.gov/rest/json/cves/2.0"
	queryParams := url.Values{}
	queryParams.Set("pubStartDate", pubStartDate)
	queryParams.Set("pubEndDate", pubEndDate)
	queryParams.Set("resultsPerPage", "100")

	req, err := http.NewRequest("GET", baseURL+"?"+queryParams.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	if s.nvdAPIKey != "" {
		req.Header.Set("apiKey", s.nvdAPIKey)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("NVD API returned status %d: %s", resp.StatusCode, string(body))
	}

	var nvdResp NVDResponse
	if err := json.NewDecoder(resp.Body).Decode(&nvdResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	items := s.parseCVEs(nvdResp.Vulnerabilities)
	return items, nil
}

func (s *CveCrawlerService) parseCVEs(vulns []NVDVulnerability) []CVEItem {
	var items []CVEItem

	for _, v := range vulns {
		severity := "UNKNOWN"
		baseScore := 0.0

		if len(v.CVE.Metrics.CvssMetricV31) > 0 {
			severity = v.CVE.Metrics.CvssMetricV31[0].CVSSData.BaseSeverity
			baseScore = v.CVE.Metrics.CvssMetricV31[0].CVSSData.BaseScore
		} else if len(v.CVE.Metrics.CvssMetricV30) > 0 {
			severity = v.CVE.Metrics.CvssMetricV30[0].CVSSData.BaseSeverity
			baseScore = v.CVE.Metrics.CvssMetricV30[0].CVSSData.BaseScore
		} else if len(v.CVE.Metrics.CvssMetricV2) > 0 {
			severity = v.CVE.Metrics.CvssMetricV2[0].CVSSData.BaseSeverity
			baseScore = v.CVE.Metrics.CvssMetricV2[0].CVSSData.BaseScore
		}

		if severity != "CRITICAL" && severity != "HIGH" {
			continue
		}

		description := ""
		for _, desc := range v.CVE.Description {
			if desc.Lang == "en" {
				description = desc.Value
				break
			}
		}
		if description == "" && len(v.CVE.Description) > 0 {
			description = v.CVE.Description[0].Value
		}

		if len(description) > 300 {
			description = description[:300] + "..."
		}

		items = append(items, CVEItem{
			ID:          v.CVE.ID,
			Severity:    severity,
			BaseScore:   baseScore,
			Description: description,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].BaseScore > items[j].BaseScore
	})

	return items
}

type CVEItem struct {
	ID          string
	Severity    string
	BaseScore   float64
	Description string
}

func (s *CveCrawlerService) formatMessageByScore(items []CVEItem, date time.Time) string {
	if len(items) == 0 {
		return ""
	}

	criticalCount := 0
	highCount := 0

	for _, item := range items {
		if item.Severity == "CRITICAL" {
			criticalCount++
		} else {
			highCount++
		}
	}

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("[info][title]🚨 DAILY CVE ALERT - %s[/title]\n", date.Format("02/01/2006")))
	sb.WriteString(fmt.Sprintf("📊 Tổng: 🔴 CRITICAL: %d | 🟠 HIGH: %d\n\n", criticalCount, highCount))

	sb.WriteString(fmt.Sprintf("🔴 CRITICAL (%d):\n", criticalCount))
	for _, item := range items {
		if item.Severity == "CRITICAL" {
			sb.WriteString(fmt.Sprintf("• %s - SCORE: %.1f\n", item.ID, item.BaseScore))
			sb.WriteString(fmt.Sprintf("  %s\n", item.Description))
			sb.WriteString(fmt.Sprintf("  🔗 https://nvd.nist.gov/vuln/detail/%s\n", item.ID))
			sb.WriteString("[hr]\n")
		}
	}

	sb.WriteString(fmt.Sprintf("🟠 HIGH (%d):\n", highCount))
	for _, item := range items {
		if item.Severity == "HIGH" {
			sb.WriteString(fmt.Sprintf("• %s - SCORE: %.1f\n", item.ID, item.BaseScore))
			sb.WriteString(fmt.Sprintf("  %s\n", item.Description))
			sb.WriteString(fmt.Sprintf("  🔗 https://nvd.nist.gov/vuln/detail/%s\n", item.ID))
			sb.WriteString("[hr]\n")
		}
	}

	sb.WriteString("📧 Powered by CVE Crawler\n")
	sb.WriteString("[/info]")

	return sb.String()
}
