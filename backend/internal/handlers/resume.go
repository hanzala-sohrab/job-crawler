package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/hanzala/job-crawler/internal/config"
	"github.com/hanzala/job-crawler/internal/models"
	"github.com/hanzala/job-crawler/internal/services/parser"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ResumeHandler struct {
	db     *pgxpool.Pool
	cfg    *config.Config
	parser parser.ResumeParser
}

func NewResumeHandler(db *pgxpool.Pool, cfg *config.Config, p parser.ResumeParser) *ResumeHandler {
	return &ResumeHandler{db: db, cfg: cfg, parser: p}
}

// Upload handles resume file upload and parsing
func (h *ResumeHandler) Upload(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(string)

	// Parse multipart form (max 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "File too large (max 10MB)")
		return
	}

	file, header, err := r.FormFile("resume")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Resume file is required (form field: 'resume')")
		return
	}
	defer file.Close()

	// Validate file type
	ext := filepath.Ext(header.Filename)
	if ext != ".pdf" && ext != ".docx" && ext != ".doc" {
		writeError(w, http.StatusBadRequest, "Only PDF and DOCX files are supported")
		return
	}

	// Save file to disk
	userDir := filepath.Join(h.cfg.UploadDir, userID)
	if err := os.MkdirAll(userDir, 0755); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create upload directory")
		return
	}

	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	filePath := filepath.Join(userDir, filename)
	dst, err := os.Create(filePath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to save file")
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to save file")
		return
	}

	// Parse resume
	parsed, rawText, err := h.parser.Parse(r.Context(), filePath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to parse resume: %v", err))
		return
	}

	parsedJSON, err := json.Marshal(parsed)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to marshal parsed data")
		return
	}

	// Upsert resume in database (one resume per user for now)
	var resume models.Resume
	err = h.db.QueryRow(r.Context(),
		`INSERT INTO resumes (user_id, file_path, file_name, raw_text, parsed_data)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (user_id) WHERE user_id = $1
		 DO UPDATE SET file_path = $2, file_name = $3, raw_text = $4, parsed_data = $5, updated_at = NOW()
		 RETURNING id, user_id, file_path, file_name, parsed_data, created_at, updated_at`,
		userID, filePath, header.Filename, rawText, parsedJSON,
	).Scan(&resume.ID, &resume.UserID, &resume.FilePath, &resume.FileName,
		&resume.ParsedData, &resume.CreatedAt, &resume.UpdatedAt)

	if err != nil {
		// Fallback: try insert without ON CONFLICT (user_id doesn't have unique constraint by default)
		err = h.db.QueryRow(r.Context(),
			`INSERT INTO resumes (user_id, file_path, file_name, raw_text, parsed_data)
			 VALUES ($1, $2, $3, $4, $5)
			 RETURNING id, user_id, file_path, file_name, parsed_data, created_at, updated_at`,
			userID, filePath, header.Filename, rawText, parsedJSON,
		).Scan(&resume.ID, &resume.UserID, &resume.FilePath, &resume.FileName,
			&resume.ParsedData, &resume.CreatedAt, &resume.UpdatedAt)

		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to save resume")
			return
		}
	}

	writeJSON(w, http.StatusCreated, resume)
}

// Get returns the user's most recent parsed resume
func (h *ResumeHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(string)

	var resume models.Resume
	err := h.db.QueryRow(r.Context(),
		`SELECT id, user_id, file_path, file_name, parsed_data, created_at, updated_at
		 FROM resumes WHERE user_id = $1 ORDER BY updated_at DESC LIMIT 1`,
		userID,
	).Scan(&resume.ID, &resume.UserID, &resume.FilePath, &resume.FileName,
		&resume.ParsedData, &resume.CreatedAt, &resume.UpdatedAt)

	if err != nil {
		writeError(w, http.StatusNotFound, "No resume found. Please upload one first.")
		return
	}

	writeJSON(w, http.StatusOK, resume)
}

// Update allows the user to edit their parsed resume data
func (h *ResumeHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(string)

	var parsedData json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&parsedData); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	var resume models.Resume
	err := h.db.QueryRow(r.Context(),
		`UPDATE resumes SET parsed_data = $1, updated_at = NOW()
		 WHERE user_id = $2
		 RETURNING id, user_id, file_path, file_name, parsed_data, created_at, updated_at`,
		parsedData, userID,
	).Scan(&resume.ID, &resume.UserID, &resume.FilePath, &resume.FileName,
		&resume.ParsedData, &resume.CreatedAt, &resume.UpdatedAt)

	if err != nil {
		writeError(w, http.StatusNotFound, "No resume found to update")
		return
	}

	writeJSON(w, http.StatusOK, resume)
}
