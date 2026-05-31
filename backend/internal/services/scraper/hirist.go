package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/hanzala/job-crawler/internal/models"
)

// HiristScraper scrapes job listings from hirist.tech using their public API
type HiristScraper struct {
	enabled bool
}

func NewHiristScraper() *HiristScraper {
	return &HiristScraper{enabled: true}
}

func (s *HiristScraper) Name() string    { return "hirist" }
func (s *HiristScraper) IsEnabled() bool { return s.enabled }

func (s *HiristScraper) Scrape(ctx context.Context, query Query) ([]models.Job, error) {
	fmt.Printf("🔍 Hirist scraper called with keywords: %v\n", query.Keywords)
	
	keywordStr := strings.Join(query.Keywords, "+")
	if keywordStr == "" {
		keywordStr = "Software+Engineer"
	}
	
	// Use the API endpoint discovered via MCP chrome-devtools
	apiURL := fmt.Sprintf("https://gladiator.hirist.tech/job/search?query=%s&page=%d&posting=0&industry=&size=20", 
		url.QueryEscape(keywordStr), query.Page-1)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/148.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", "https://www.hirist.tech/")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("hirist API returned non-200 status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse JSON Response
	var res struct {
		Data []struct {
			ID          int    `json:"id"`
			Title       string `json:"title"`
			MinExp      int    `json:"min"`
			MaxExp      int    `json:"max"`
			JobURL      string `json:"jobDetailUrl"`
			CreatedTime int64  `json:"createdTime"`
			Tags        []struct {
				Name string `json:"name"`
			} `json:"tags"`
			Locations []struct {
				Name string `json:"name"`
			} `json:"locations"`
			CompanyData struct {
				CompanyName string `json:"companyName"`
			} `json:"companyData"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	var jobs []models.Job
	for _, item := range res.Data {
		var skills []string
		for _, tag := range item.Tags {
			skills = append(skills, tag.Name)
		}
		skillsJSON, _ := json.Marshal(skills)

		var locations []string
		for _, loc := range item.Locations {
			locations = append(locations, loc.Name)
		}

		postedAt := time.UnixMilli(item.CreatedTime)

		jobs = append(jobs, models.Job{
			ExternalID:      strconv.Itoa(item.ID),
			Source:          s.Name(),
			Title:           item.Title,
			Company:         item.CompanyData.CompanyName,
			Location:        strings.Join(locations, ", "),
			Description:     "View job detail page for more information.",
			Requirements:    "See job link for requirements.",
			JobType:         "Full-time",
			ExperienceLevel: fmt.Sprintf("%d-%d years", item.MinExp, item.MaxExp),
			Skills:          json.RawMessage(skillsJSON),
			URL:             item.JobURL,
			PostedAt:        &postedAt,
			ScrapedAt:       time.Now(),
			IsActive:        true,
		})
	}

	fmt.Printf("✅ Hirist scraper successfully fetched %d jobs via API\n", len(jobs))
	return jobs, nil
}
