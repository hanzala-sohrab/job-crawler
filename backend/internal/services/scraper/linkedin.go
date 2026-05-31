package scraper

import (
	"context"
	"fmt"

	"github.com/hanzala/job-crawler/internal/models"
)

// LinkedInScraper scrapes job listings from LinkedIn
type LinkedInScraper struct {
	enabled bool
}

func NewLinkedInScraper() *LinkedInScraper {
	return &LinkedInScraper{enabled: false} // disabled by default (needs chromedp)
}

func (s *LinkedInScraper) Name() string    { return "linkedin" }
func (s *LinkedInScraper) IsEnabled() bool { return s.enabled }

func (s *LinkedInScraper) Scrape(ctx context.Context, query Query) ([]models.Job, error) {
	// TODO: Implement with Chromedp for JS rendering
	fmt.Printf("🔍 LinkedIn scraper called with keywords: %v\n", query.Keywords)
	return []models.Job{}, nil
}
