package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/hanzala/job-crawler/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	db       *pgxpool.Pool
	scrapers []Scraper
	cron     *cron.Cron
	mu       sync.Mutex
}

func NewScheduler(db *pgxpool.Pool) *Scheduler {
	return &Scheduler{db: db, cron: cron.New()}
}

func (s *Scheduler) RegisterScraper(sc Scraper) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sc.IsEnabled() {
		s.scrapers = append(s.scrapers, sc)
		fmt.Printf("📋 Registered scraper: %s\n", sc.Name())
	}
}

func (s *Scheduler) Start(cronExpr string) error {
	_, err := s.cron.AddFunc(cronExpr, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		s.RunAll(ctx)
	})
	if err != nil {
		return fmt.Errorf("failed to schedule: %w", err)
	}
	s.cron.Start()
	fmt.Printf("🕐 Scheduler started: %s\n", cronExpr)
	return nil
}

func (s *Scheduler) Stop() { s.cron.Stop() }

func (s *Scheduler) RunAll(ctx context.Context) {
	s.mu.Lock()
	scrapers := make([]Scraper, len(s.scrapers))
	copy(scrapers, s.scrapers)
	s.mu.Unlock()

	var wg sync.WaitGroup
	for _, sc := range scrapers {
		wg.Add(1)
		go func(sc Scraper) {
			defer wg.Done()
			s.runScraper(ctx, sc)
		}(sc)
	}
	wg.Wait()
}

func (s *Scheduler) runScraper(ctx context.Context, sc Scraper) {
	query := Query{
		Keywords: []string{"software engineer", "backend developer", "node.js", "next.js", "react.js", "javascript", "frontend developer", "fullstack developer", "mern stack developer", "django", "python", "golang", "typescript", "redux"},
		Page:     1,
	}
	jobs, err := sc.Scrape(ctx, query)
	if err != nil {
		fmt.Printf("❌ %s failed: %v\n", sc.Name(), err)
		return
	}
	inserted := 0
	for _, job := range jobs {
		skillsJSON, _ := json.Marshal(job.Skills)
		if skillsJSON == nil {
			skillsJSON = []byte("[]")
		}
		_, err := s.db.Exec(ctx,
			`INSERT INTO jobs (external_id, source, title, company, location, description,
				requirements, salary_range, job_type, experience_level, skills, url, posted_at)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
			 ON CONFLICT (source, external_id) DO UPDATE SET
				title=EXCLUDED.title, description=EXCLUDED.description,
				salary_range=EXCLUDED.salary_range, skills=EXCLUDED.skills, scraped_at=NOW()`,
			job.ExternalID, sc.Name(), job.Title, job.Company, job.Location,
			job.Description, job.Requirements, job.SalaryRange, job.JobType,
			job.ExperienceLevel, skillsJSON, job.URL, job.PostedAt,
		)
		if err == nil {
			inserted++
		}
	}
	fmt.Printf("✅ %s: %d found, %d saved\n", sc.Name(), len(jobs), inserted)
}

// Ensure models is used
var _ = models.Job{}
