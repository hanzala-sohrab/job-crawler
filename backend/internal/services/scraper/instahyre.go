package scraper

import (
	"context"
	"fmt"

	"github.com/hanzala/job-crawler/internal/models"
)

// InstahyreScraper scrapes job listings from Instahyre
type InstahyreScraper struct {
	enabled bool
}

func NewInstahyreScraper() *InstahyreScraper {
	return &InstahyreScraper{enabled: false}
}

func (s *InstahyreScraper) Name() string    { return "instahyre" }
func (s *InstahyreScraper) IsEnabled() bool { return s.enabled }

func (s *InstahyreScraper) Scrape(ctx context.Context, query Query) ([]models.Job, error) {
	fmt.Printf("🔍 Instahyre scraper called with keywords: %v\n", query.Keywords)
	return []models.Job{}, nil
}
