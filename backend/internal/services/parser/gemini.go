package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hanzala/job-crawler/internal/models"
	"github.com/hanzala/job-crawler/internal/utils"
	"google.golang.org/genai"
)

// GeminiParser uses Google's Gemini API for structured resume extraction
type GeminiParser struct {
	client    *genai.Client
	modelName string
}

// NewGeminiParser creates a new Gemini-based resume parser
func NewGeminiParser(ctx context.Context, apiKey string, modelName string) (*GeminiParser, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &GeminiParser{
		client:    client,
		modelName: modelName,
	}, nil
}

func (p *GeminiParser) Name() string {
	return "gemini"
}

// Parse extracts text from a resume file and uses Gemini to structure it
func (p *GeminiParser) Parse(ctx context.Context, filePath string) (*models.ParsedResume, string, error) {
	// Step 1: Extract raw text from the file
	rawText, err := extractText(filePath)
	if err != nil {
		return nil, "", fmt.Errorf("text extraction failed: %w", err)
	}

	if strings.TrimSpace(rawText) == "" {
		return nil, "", fmt.Errorf("no text could be extracted from the file")
	}

	// Step 2: Send to Gemini with structured output schema
	schema := &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"full_name":              {Type: genai.TypeString},
			"email":                  {Type: genai.TypeString},
			"phone":                  {Type: genai.TypeString},
			"location":               {Type: genai.TypeString},
			"summary":                {Type: genai.TypeString},
			"total_experience_years": {Type: genai.TypeNumber},
			"skills": {
				Type:  genai.TypeArray,
				Items: &genai.Schema{Type: genai.TypeString},
			},
			"experience": {
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"company":     {Type: genai.TypeString},
						"title":       {Type: genai.TypeString},
						"start_date":  {Type: genai.TypeString},
						"end_date":    {Type: genai.TypeString},
						"description": {Type: genai.TypeString},
						"technologies": {
							Type:  genai.TypeArray,
							Items: &genai.Schema{Type: genai.TypeString},
						},
					},
				},
			},
			"education": {
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"institution": {Type: genai.TypeString},
						"degree":      {Type: genai.TypeString},
						"field":       {Type: genai.TypeString},
						"year":        {Type: genai.TypeInteger},
					},
				},
			},
			"certifications": {
				Type:  genai.TypeArray,
				Items: &genai.Schema{Type: genai.TypeString},
			},
			"preferred_job_titles": {
				Type:  genai.TypeArray,
				Items: &genai.Schema{Type: genai.TypeString},
			},
			"preferred_locations": {
				Type:  genai.TypeArray,
				Items: &genai.Schema{Type: genai.TypeString},
			},
		},
		Required: []string{"full_name", "skills"},
	}

	temperature := float32(0.1)
	config := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		ResponseSchema:   schema,
		Temperature:      &temperature,
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{
				genai.NewPartFromText(
					`You are an expert resume parser. Extract structured information from the resume text provided.
Be thorough in extracting ALL skills mentioned — including programming languages, frameworks, tools, and soft skills.
For preferred_job_titles, infer 3-5 job titles that match the candidate's experience and skills.
For preferred_locations, extract any mentioned location preferences; if none, leave empty.
For total_experience_years, calculate the total years of professional experience.`),
			},
		},
	}

	contents := []*genai.Content{
		{
			Parts: []*genai.Part{
				genai.NewPartFromText("Parse the following resume and extract structured data:\n\n" + rawText),
			},
			Role: "user",
		},
	}

	resp, err := p.client.Models.GenerateContent(ctx, p.modelName, contents, config)
	if err != nil {
		return nil, rawText, fmt.Errorf("Gemini API call failed: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, rawText, fmt.Errorf("empty response from Gemini")
	}

	responseText := resp.Candidates[0].Content.Parts[0].Text

	// Step 3: Parse the JSON response
	var parsed models.ParsedResume
	if err := json.Unmarshal([]byte(responseText), &parsed); err != nil {
		return nil, rawText, fmt.Errorf("failed to parse Gemini response: %w", err)
	}

	return &parsed, rawText, nil
}

// extractText reads text content from PDF or DOCX files
func extractText(filePath string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".pdf":
		return utils.ExtractPDFText(filePath)
	case ".docx", ".doc":
		return utils.ExtractDOCXText(filePath)
	default:
		// Try reading as plain text
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}
}
