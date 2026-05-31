package parser

import (
	"context"

	"github.com/hanzala/job-crawler/internal/models"
)

// ResumeParser is the interface for parsing resumes into structured data
type ResumeParser interface {
	// Parse takes a file path, extracts text, and returns structured data
	Parse(ctx context.Context, filePath string) (*models.ParsedResume, string, error)
	// Name returns the parser implementation name
	Name() string
}
