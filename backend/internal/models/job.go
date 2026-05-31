package models

import (
	"encoding/json"
	"time"
)

// Job represents a scraped job listing
type Job struct {
	ID              string          `json:"id"`
	ExternalID      string          `json:"external_id,omitempty"`
	Source          string          `json:"source"`
	Title           string          `json:"title"`
	Company         string          `json:"company"`
	Location        string          `json:"location"`
	Description     string          `json:"description"`
	Requirements    string          `json:"requirements,omitempty"`
	SalaryRange     string          `json:"salary_range,omitempty"`
	JobType         string          `json:"job_type,omitempty"`
	ExperienceLevel string          `json:"experience_level,omitempty"`
	Skills          json.RawMessage `json:"skills"`
	URL             string          `json:"url"`
	PostedAt        *time.Time      `json:"posted_at,omitempty"`
	ScrapedAt       time.Time       `json:"scraped_at"`
	IsActive        bool            `json:"is_active"`
	CreatedAt       time.Time       `json:"created_at"`
}

// JobMatch represents a match between a user's resume and a job listing
type JobMatch struct {
	ID             string          `json:"id"`
	UserID         string          `json:"user_id"`
	JobID          string          `json:"job_id"`
	RelevanceScore float64         `json:"relevance_score"`
	MatchedSkills  json.RawMessage `json:"matched_skills"`
	Status         string          `json:"status"` // new, saved, applied, hidden
	CreatedAt      time.Time       `json:"created_at"`
	// Embedded job data for list responses
	Job *Job `json:"job,omitempty"`
}

// JobFilter represents query parameters for job listing
type JobFilter struct {
	Source          string   `json:"source,omitempty"`
	Skills          []string `json:"skills,omitempty"`
	Location        string   `json:"location,omitempty"`
	JobType         string   `json:"job_type,omitempty"`
	ExperienceLevel string   `json:"experience_level,omitempty"`
	Status          string   `json:"status,omitempty"`
	Query           string   `json:"query,omitempty"`
	Page            int      `json:"page"`
	PageSize        int      `json:"page_size"`
	SortBy          string   `json:"sort_by,omitempty"` // relevance, posted_at, created_at
}

// PaginatedResponse wraps a list response with pagination metadata
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}
