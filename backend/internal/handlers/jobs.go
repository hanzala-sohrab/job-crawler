package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/hanzala/job-crawler/internal/models"
	"github.com/hanzala/job-crawler/internal/services/matcher"
	"github.com/jackc/pgx/v5/pgxpool"
)

type JobHandler struct {
	db      *pgxpool.Pool
	matcher *matcher.Matcher
}

func NewJobHandler(db *pgxpool.Pool, m *matcher.Matcher) *JobHandler {
	return &JobHandler{db: db, matcher: m}
}

// List returns matched jobs for the authenticated user with filters
func (h *JobHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(string)

	filter := models.JobFilter{
		Source:          r.URL.Query().Get("source"),
		Location:        r.URL.Query().Get("location"),
		JobType:         r.URL.Query().Get("job_type"),
		ExperienceLevel: r.URL.Query().Get("experience_level"),
		Status:          r.URL.Query().Get("status"),
		Query:           r.URL.Query().Get("q"),
		SortBy:          r.URL.Query().Get("sort_by"),
	}

	// Skills from comma-separated query param
	if skills := r.URL.Query().Get("skills"); skills != "" {
		filter.Skills = strings.Split(skills, ",")
	}

	// Pagination
	filter.Page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	if filter.Page < 1 {
		filter.Page = 1
	}
	filter.PageSize, _ = strconv.Atoi(r.URL.Query().Get("page_size"))
	if filter.PageSize < 1 || filter.PageSize > 50 {
		filter.PageSize = 20
	}

	if filter.SortBy == "" {
		filter.SortBy = "relevance"
	}

	offset := (filter.Page - 1) * filter.PageSize

	// Build dynamic query
	query := `
		SELECT jm.id, jm.relevance_score, jm.matched_skills, jm.status, jm.created_at,
			   j.id, j.source, j.title, j.company, j.location, j.description,
			   j.salary_range, j.job_type, j.experience_level, j.skills, j.url, 
			   j.posted_at, j.is_active
		FROM job_matches jm
		JOIN jobs j ON j.id = jm.job_id
		WHERE jm.user_id = $1 AND j.is_active = true`

	countQuery := `
		SELECT COUNT(*)
		FROM job_matches jm
		JOIN jobs j ON j.id = jm.job_id
		WHERE jm.user_id = $1 AND j.is_active = true`

	args := []interface{}{userID}
	argIdx := 2

	if filter.Source != "" {
		query += ` AND j.source = $` + strconv.Itoa(argIdx)
		countQuery += ` AND j.source = $` + strconv.Itoa(argIdx)
		args = append(args, filter.Source)
		argIdx++
	}
	if filter.Location != "" {
		query += ` AND j.location ILIKE $` + strconv.Itoa(argIdx)
		countQuery += ` AND j.location ILIKE $` + strconv.Itoa(argIdx)
		args = append(args, "%"+filter.Location+"%")
		argIdx++
	}
	if filter.JobType != "" {
		query += ` AND j.job_type = $` + strconv.Itoa(argIdx)
		countQuery += ` AND j.job_type = $` + strconv.Itoa(argIdx)
		args = append(args, filter.JobType)
		argIdx++
	}
	if filter.ExperienceLevel != "" {
		query += ` AND j.experience_level = $` + strconv.Itoa(argIdx)
		countQuery += ` AND j.experience_level = $` + strconv.Itoa(argIdx)
		args = append(args, filter.ExperienceLevel)
		argIdx++
	}
	if filter.Status != "" {
		query += ` AND jm.status = $` + strconv.Itoa(argIdx)
		countQuery += ` AND jm.status = $` + strconv.Itoa(argIdx)
		args = append(args, filter.Status)
		argIdx++
	}
	if filter.Query != "" {
		query += ` AND (j.title ILIKE $` + strconv.Itoa(argIdx) +
			` OR j.company ILIKE $` + strconv.Itoa(argIdx) +
			` OR j.description ILIKE $` + strconv.Itoa(argIdx) + `)`
		countQuery += ` AND (j.title ILIKE $` + strconv.Itoa(argIdx) +
			` OR j.company ILIKE $` + strconv.Itoa(argIdx) +
			` OR j.description ILIKE $` + strconv.Itoa(argIdx) + `)`
		args = append(args, "%"+filter.Query+"%")
		argIdx++
	}

	// Sorting
	switch filter.SortBy {
	case "posted_at":
		query += ` ORDER BY j.posted_at DESC NULLS LAST`
	case "created_at":
		query += ` ORDER BY jm.created_at DESC`
	default:
		query += ` ORDER BY jm.relevance_score DESC`
	}

	// Get total count
	var total int
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	h.db.QueryRow(r.Context(), countQuery, countArgs...).Scan(&total)

	// Pagination
	query += ` LIMIT $` + strconv.Itoa(argIdx) + ` OFFSET $` + strconv.Itoa(argIdx+1)
	args = append(args, filter.PageSize, offset)

	rows, err := h.db.Query(r.Context(), query, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to query jobs")
		return
	}
	defer rows.Close()

	var matches []models.JobMatch
	for rows.Next() {
		var m models.JobMatch
		var j models.Job
		err := rows.Scan(
			&m.ID, &m.RelevanceScore, &m.MatchedSkills, &m.Status, &m.CreatedAt,
			&j.ID, &j.Source, &j.Title, &j.Company, &j.Location, &j.Description,
			&j.SalaryRange, &j.JobType, &j.ExperienceLevel, &j.Skills, &j.URL,
			&j.PostedAt, &j.IsActive,
		)
		if err != nil {
			fmt.Printf("❌ rows.Scan error in List jobs: %v\n", err)
			continue
		}
		m.Job = &j
		matches = append(matches, m)
	}

	if matches == nil {
		matches = []models.JobMatch{}
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data:       matches,
		Total:      total,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: int(math.Ceil(float64(total) / float64(filter.PageSize))),
	})
}

// GetByID returns a specific job's details
func (h *JobHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "id")

	var job models.Job
	err := h.db.QueryRow(r.Context(),
		`SELECT id, external_id, source, title, company, location, description,
			    requirements, salary_range, job_type, experience_level, skills,
			    url, posted_at, scraped_at, is_active, created_at
		 FROM jobs WHERE id = $1`,
		jobID,
	).Scan(&job.ID, &job.ExternalID, &job.Source, &job.Title, &job.Company,
		&job.Location, &job.Description, &job.Requirements, &job.SalaryRange,
		&job.JobType, &job.ExperienceLevel, &job.Skills, &job.URL,
		&job.PostedAt, &job.ScrapedAt, &job.IsActive, &job.CreatedAt)

	if err != nil {
		writeError(w, http.StatusNotFound, "Job not found")
		return
	}

	writeJSON(w, http.StatusOK, job)
}

// UpdateStatus updates the status of a job match (saved, applied, hidden)
func (h *JobHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(string)
	jobID := chi.URLParam(r, "id")

	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	validStatuses := map[string]bool{"new": true, "saved": true, "applied": true, "hidden": true}
	if !validStatuses[body.Status] {
		writeError(w, http.StatusBadRequest, "Invalid status. Must be: new, saved, applied, or hidden")
		return
	}

	tag, err := h.db.Exec(r.Context(),
		`UPDATE job_matches SET status = $1 WHERE user_id = $2 AND job_id = $3`,
		body.Status, userID, jobID,
	)

	if err != nil || tag.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "Job match not found")
		return
	}

	writeMessage(w, http.StatusOK, "Status updated")
}

// Sources returns a list of active job sources
func (h *JobHandler) Sources(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(r.Context(),
		`SELECT source, COUNT(*) as count 
		 FROM jobs WHERE is_active = true 
		 GROUP BY source ORDER BY count DESC`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to query sources")
		return
	}
	defer rows.Close()

	type SourceInfo struct {
		Source string `json:"source"`
		Count  int    `json:"count"`
	}

	var sources []SourceInfo
	for rows.Next() {
		var s SourceInfo
		if err := rows.Scan(&s.Source, &s.Count); err != nil {
			continue
		}
		sources = append(sources, s)
	}

	if sources == nil {
		sources = []SourceInfo{}
	}

	writeJSON(w, http.StatusOK, sources)
}

// RefreshMatches triggers re-matching for the current user
func (h *JobHandler) RefreshMatches(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(string)

	// Get user's resume
	var parsedData json.RawMessage
	err := h.db.QueryRow(r.Context(),
		`SELECT parsed_data FROM resumes WHERE user_id = $1 ORDER BY updated_at DESC LIMIT 1`,
		userID,
	).Scan(&parsedData)

	if err != nil {
		writeError(w, http.StatusBadRequest, "No resume found. Please upload a resume first.")
		return
	}

	var parsed models.ParsedResume
	if err := json.Unmarshal(parsedData, &parsed); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to parse resume data")
		return
	}

	// Run matching in the background using a new context, as r.Context() will be
	// canceled immediately when this handler returns.
	go func() {
		bgCtx := context.Background()
		if err := h.matcher.MatchJobsForUser(bgCtx, userID, &parsed); err != nil {
			fmt.Printf("❌ Failed to match jobs for user %s: %v\n", userID, err)
		}
	}()

	writeMessage(w, http.StatusAccepted, "Job matching started. Refresh in a few seconds to see results.")
}

// Need fmt for RefreshMatches goroutine
func init() {
	_ = json.RawMessage{}
}
