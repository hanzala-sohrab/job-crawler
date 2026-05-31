package scraper

import (
	"context"

	"github.com/hanzala/job-crawler/internal/models"
)

// Scraper is the interface that all job source scrapers must implement
type Scraper interface {
	// Name returns the scraper source name (e.g., "naukri", "linkedin")
	Name() string
	// Scrape fetches job listings matching the given query
	Scrape(ctx context.Context, query Query) ([]models.Job, error)
	// IsEnabled returns whether this scraper is configured and ready
	IsEnabled() bool
}

// Query represents search parameters for scrapers
type Query struct {
	Keywords   []string `json:"keywords"`
	Location   string   `json:"location"`
	Experience string   `json:"experience"` // e.g., "0-2", "3-5", "5-10", "10+"
	JobType    string   `json:"job_type"`    // "full-time", "remote", "contract"
	Page       int      `json:"page"`
}
