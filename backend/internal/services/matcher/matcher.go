package matcher

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/hanzala/job-crawler/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Matcher computes relevance scores between resumes and jobs
type Matcher struct {
	db *pgxpool.Pool
}

func NewMatcher(db *pgxpool.Pool) *Matcher {
	return &Matcher{db: db}
}

// Weights for scoring components
const (
	weightSkills     = 0.40
	weightTitle      = 0.25
	weightExperience = 0.20
	weightLocation   = 0.15
)

// MatchJobsForUser scores all active jobs against the user's parsed resume
func (m *Matcher) MatchJobsForUser(ctx context.Context, userID string, resume *models.ParsedResume) error {
	// Fetch all active jobs
	rows, err := m.db.Query(ctx,
		`SELECT id, title, company, location, experience_level, skills 
		 FROM jobs WHERE is_active = true`)
	if err != nil {
		return fmt.Errorf("failed to fetch jobs: %w", err)
	}
	defer rows.Close()

	type jobRow struct {
		ID              string
		Title           string
		Company         string
		Location        string
		ExperienceLevel string
		Skills          json.RawMessage
	}

	var jobs []jobRow
	for rows.Next() {
		var j jobRow
		if err := rows.Scan(&j.ID, &j.Title, &j.Company, &j.Location, &j.ExperienceLevel, &j.Skills); err != nil {
			continue
		}
		jobs = append(jobs, j)
	}

	// Normalize resume skills to lowercase
	resumeSkills := make(map[string]bool)
	for _, s := range resume.Skills {
		resumeSkills[strings.ToLower(strings.TrimSpace(s))] = true
	}

	resumeTitles := resume.PreferredJobTitles
	
	// Consider user's actual work experience to boost matching
	for _, exp := range resume.Experience {
		if exp.Title != "" {
			resumeTitles = append(resumeTitles, exp.Title)
		}
		for _, tech := range exp.Technologies {
			resumeSkills[strings.ToLower(strings.TrimSpace(tech))] = true
		}
	}

	resumeLocations := resume.PreferredLocations
	resumeYears := resume.TotalExperienceYears

	// Score each job
	for _, job := range jobs {
		score, matchedSkills := scoreJob(
			resumeSkills, resumeTitles, resumeLocations, resumeYears,
			job.Title, job.Location, job.ExperienceLevel, job.Skills,
		)

		// Only create matches with meaningful scores
		if score < 0.05 {
			continue
		}

		matchedJSON, _ := json.Marshal(matchedSkills)

		// Upsert the match
		_, err := m.db.Exec(ctx,
			`INSERT INTO job_matches (user_id, job_id, relevance_score, matched_skills)
			 VALUES ($1, $2, $3, $4)
			 ON CONFLICT (user_id, job_id)
			 DO UPDATE SET relevance_score = $3, matched_skills = $4`,
			userID, job.ID, score, matchedJSON,
		)
		if err != nil {
			fmt.Printf("⚠️  Failed to upsert match for job %s: %v\n", job.ID, err)
		}
	}

	fmt.Printf("✅ Matched %d jobs for user %s\n", len(jobs), userID)
	return nil
}

func scoreJob(
	resumeSkills map[string]bool,
	resumeTitles []string,
	resumeLocations []string,
	resumeYears float64,
	jobTitle, jobLocation, jobExpLevel string,
	jobSkillsRaw json.RawMessage,
) (float64, []string) {
	// 1. Skill Match (Jaccard-like)
	var jobSkills []string
	json.Unmarshal(jobSkillsRaw, &jobSkills)

	var matchedSkills []string
	for _, js := range jobSkills {
		if resumeSkills[strings.ToLower(strings.TrimSpace(js))] {
			matchedSkills = append(matchedSkills, js)
		}
	}

	skillScore := 0.0
	if len(jobSkills) > 0 {
		skillScore = float64(len(matchedSkills)) / float64(len(jobSkills))
	}

	// 2. Title Match
	titleScore := 0.0
	jobTitleLower := strings.ToLower(jobTitle)
	for _, t := range resumeTitles {
		sim := stringSimilarity(strings.ToLower(t), jobTitleLower)
		if sim > titleScore {
			titleScore = sim
		}
	}

	// 3. Experience Match
	expScore := scoreExperience(resumeYears, jobExpLevel)

	// 4. Location Match
	locScore := 0.0
	jobLocLower := strings.ToLower(jobLocation)
	if strings.Contains(jobLocLower, "remote") {
		locScore = 1.0
	} else {
		for _, loc := range resumeLocations {
			if strings.Contains(jobLocLower, strings.ToLower(loc)) ||
				strings.Contains(strings.ToLower(loc), jobLocLower) {
				locScore = 1.0
				break
			}
		}
	}
	// If no preferred locations set, don't penalize
	if len(resumeLocations) == 0 {
		locScore = 0.5
	}

	total := skillScore*weightSkills +
		titleScore*weightTitle +
		expScore*weightExperience +
		locScore*weightLocation

	// Clamp to [0, 1]
	total = math.Min(1.0, math.Max(0.0, total))

	return total, matchedSkills
}

func scoreExperience(resumeYears float64, jobExpLevel string) float64 {
	switch strings.ToLower(jobExpLevel) {
	case "entry", "junior", "fresher":
		if resumeYears <= 2 {
			return 1.0
		}
		return 0.5
	case "mid", "intermediate":
		if resumeYears >= 2 && resumeYears <= 6 {
			return 1.0
		}
		if resumeYears < 2 {
			return 0.3
		}
		return 0.7
	case "senior":
		if resumeYears >= 5 {
			return 1.0
		}
		return 0.3
	case "lead", "principal", "staff":
		if resumeYears >= 8 {
			return 1.0
		}
		return 0.2
	default:
		return 0.5 // Unknown level, neutral score
	}
}

// stringSimilarity computes a simple word-overlap similarity between two strings
func stringSimilarity(a, b string) float64 {
	wordsA := strings.Fields(a)
	wordsB := strings.Fields(b)

	if len(wordsA) == 0 || len(wordsB) == 0 {
		return 0
	}

	setB := make(map[string]bool)
	for _, w := range wordsB {
		setB[w] = true
	}

	matches := 0
	for _, w := range wordsA {
		if setB[w] {
			matches++
		}
	}

	// Dice coefficient
	return 2.0 * float64(matches) / float64(len(wordsA)+len(wordsB))
}
