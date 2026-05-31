package models

import (
	"encoding/json"
	"time"
)

// Resume represents a user's uploaded and parsed resume
type Resume struct {
	ID         string          `json:"id"`
	UserID     string          `json:"user_id"`
	FilePath   string          `json:"file_path,omitempty"`
	FileName   string          `json:"file_name,omitempty"`
	RawText    string          `json:"raw_text,omitempty"`
	ParsedData json.RawMessage `json:"parsed_data,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

// ParsedResume is the structured data extracted from a resume by AI
type ParsedResume struct {
	FullName             string       `json:"full_name"`
	Email                string       `json:"email"`
	Phone                string       `json:"phone"`
	Location             string       `json:"location"`
	Summary              string       `json:"summary"`
	TotalExperienceYears float64      `json:"total_experience_years"`
	Skills               []string     `json:"skills"`
	Experience           []Experience `json:"experience"`
	Education            []Education  `json:"education"`
	Certifications       []string     `json:"certifications"`
	PreferredJobTitles   []string     `json:"preferred_job_titles"`
	PreferredLocations   []string     `json:"preferred_locations"`
}

// Experience represents a work experience entry
type Experience struct {
	Company      string   `json:"company"`
	Title        string   `json:"title"`
	StartDate    string   `json:"start_date"`
	EndDate      string   `json:"end_date"`
	Description  string   `json:"description"`
	Technologies []string `json:"technologies"`
}

// Education represents an education entry
type Education struct {
	Institution string `json:"institution"`
	Degree      string `json:"degree"`
	Field       string `json:"field"`
	Year        int    `json:"year"`
}
