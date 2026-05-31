package parser

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hanzala/job-crawler/internal/models"
)

// LocalParser is a regex/heuristic-based fallback parser that doesn't require an API key
type LocalParser struct{}

func NewLocalParser() *LocalParser {
	return &LocalParser{}
}

func (p *LocalParser) Name() string {
	return "local"
}

// Parse extracts text and uses heuristics to structure it
func (p *LocalParser) Parse(ctx context.Context, filePath string) (*models.ParsedResume, string, error) {
	rawText, err := extractText(filePath)
	if err != nil {
		return nil, "", fmt.Errorf("text extraction failed: %w", err)
	}

	if strings.TrimSpace(rawText) == "" {
		return nil, "", fmt.Errorf("no text could be extracted from the file")
	}

	parsed := &models.ParsedResume{}

	// Extract email
	emailRegex := regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)
	if matches := emailRegex.FindStringSubmatch(rawText); len(matches) > 0 {
		parsed.Email = matches[0]
	}

	// Extract phone
	phoneRegex := regexp.MustCompile(`[\+]?[(]?[0-9]{1,4}[)]?[-\s\./0-9]{7,15}`)
	if matches := phoneRegex.FindStringSubmatch(rawText); len(matches) > 0 {
		parsed.Phone = strings.TrimSpace(matches[0])
	}

	// Extract skills (common tech skills)
	knownSkills := []string{
		"go", "golang", "python", "java", "javascript", "typescript", "react", "angular",
		"vue", "node.js", "nodejs", "express", "django", "flask", "spring", "kubernetes",
		"docker", "aws", "gcp", "azure", "terraform", "jenkins", "git", "linux",
		"postgresql", "mysql", "mongodb", "redis", "elasticsearch", "kafka", "rabbitmq",
		"graphql", "rest", "grpc", "microservices", "ci/cd", "agile", "scrum",
		"html", "css", "sass", "webpack", "next.js", "nextjs", "tailwind",
		"rust", "c++", "c#", ".net", "swift", "kotlin", "flutter", "dart",
		"machine learning", "deep learning", "nlp", "computer vision", "data science",
		"sql", "nosql", "api", "devops", "sre", "system design",
	}

	lowerText := strings.ToLower(rawText)
	for _, skill := range knownSkills {
		if strings.Contains(lowerText, strings.ToLower(skill)) {
			parsed.Skills = append(parsed.Skills, skill)
		}
	}

	// Try to extract name (first line that looks like a name)
	lines := strings.Split(rawText, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		// Simple heuristic: first non-empty line with 2-4 words is likely the name
		words := strings.Fields(trimmed)
		if len(words) >= 2 && len(words) <= 4 && !strings.Contains(trimmed, "@") {
			allAlpha := true
			for _, word := range words {
				if !isAlphaString(word) {
					allAlpha = false
					break
				}
			}
			if allAlpha {
				parsed.FullName = trimmed
				break
			}
		}
	}

	return parsed, rawText, nil
}

func isAlphaString(s string) bool {
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '.' || r == '-' || r == '\'') {
			return false
		}
	}
	return len(s) > 0
}
