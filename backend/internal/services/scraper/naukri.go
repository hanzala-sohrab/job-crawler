package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
	"html"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/hanzala/job-crawler/internal/models"
)

// NaukriScraper scrapes job listings from naukri.com
type NaukriScraper struct {
	enabled bool
}

func NewNaukriScraper() *NaukriScraper {
	return &NaukriScraper{enabled: true}
}

func (s *NaukriScraper) Name() string    { return "naukri" }
func (s *NaukriScraper) IsEnabled() bool { return s.enabled }

type NaukriResponse struct {
	JobDetails []struct {
		Title          string `json:"title"`
		CompanyName    string `json:"companyName"`
		TagsAndSkills  string `json:"tagsAndSkills"`
		JobID          string `json:"jobId"`
		JobDescription string `json:"jobDescription"`
		CreatedDate    int64  `json:"createdDate"`
		JdURL          string `json:"jdURL"`
		Placeholders   []struct {
			Type  string `json:"type"`
			Label string `json:"label"`
		} `json:"placeholders"`
	} `json:"jobDetails"`
}

func (s *NaukriScraper) Scrape(ctx context.Context, query Query) ([]models.Job, error) {
	fmt.Printf("🔍 Naukri scraper called with keywords: %v\n", query.Keywords)
	
	keywordStr := strings.Join(query.Keywords, " ")
	if keywordStr == "" {
		keywordStr = "software engineer"
	}
	
	searchURL := fmt.Sprintf("https://www.naukri.com/%s-jobs", url.QueryEscape(strings.ReplaceAll(keywordStr, " ", "-")))

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.WindowSize(1920, 1080),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)
	
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(ctx, opts...)
	defer cancelAlloc()

	taskCtx, cancelTask := chromedp.NewContext(allocCtx)
	defer cancelTask()

	taskCtx, cancelTimeout := context.WithTimeout(taskCtx, 45*time.Second)
	defer cancelTimeout()

	var requestID network.RequestID
	var responseBody string
	var errScrape error

	chromedp.ListenTarget(taskCtx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *network.EventResponseReceived:
			if strings.Contains(ev.Response.URL, "/jobapi/v3/search") {
				requestID = ev.RequestID
			}
		case *network.EventLoadingFinished:
			if ev.RequestID == requestID && requestID != "" {
				go func() {
					c := chromedp.FromContext(taskCtx)
					body, err := network.GetResponseBody(ev.RequestID).Do(cdp.WithExecutor(taskCtx, c.Target))
					if err == nil {
						responseBody = string(body)
					}
				}()
			}
		}
	})

	err := chromedp.Run(taskCtx,
		chromedp.Navigate(searchURL),
		chromedp.Sleep(6*time.Second), // Wait enough time for the API call to complete
	)

	if err != nil {
		fmt.Printf("⚠️ Naukri scraper chromedp run error: %v\n", err)
		return nil, fmt.Errorf("failed to run chromedp: %w", err)
	}

	if responseBody == "" {
		fmt.Println("⚠️ Naukri scraper failed to intercept JSON response")
		return nil, fmt.Errorf("failed to intercept JSON response")
	}

	var resp NaukriResponse
	if err := json.Unmarshal([]byte(responseBody), &resp); err != nil {
		fmt.Printf("⚠️ Naukri scraper JSON parse error: %v\n", err)
		return nil, fmt.Errorf("failed to parse json: %w", err)
	}

	var jobs []models.Job
	for _, item := range resp.JobDetails {
		var exp, salary, loc string
		for _, ph := range item.Placeholders {
			switch ph.Type {
			case "experience":
				exp = ph.Label
			case "salary":
				salary = ph.Label
			case "location":
				loc = ph.Label
			}
		}

		skillsList := strings.Split(item.TagsAndSkills, ",")
		var cleanSkills []string
		for _, s := range skillsList {
			s = strings.TrimSpace(s)
			if s != "" {
				cleanSkills = append(cleanSkills, s)
			}
		}
		
		skillsJSON, _ := json.Marshal(cleanSkills)
		
		desc := strings.ReplaceAll(item.JobDescription, "<br>", "\n")
		desc = html.UnescapeString(desc)

		jobURL := item.JdURL
		if !strings.HasPrefix(jobURL, "http") {
			jobURL = "https://www.naukri.com" + jobURL
		}

		postedAt := time.UnixMilli(item.CreatedDate)

		jobs = append(jobs, models.Job{
			ExternalID:      item.JobID,
			Source:          s.Name(),
			Title:           html.UnescapeString(item.Title),
			Company:         html.UnescapeString(item.CompanyName),
			Location:        html.UnescapeString(loc),
			Description:     desc,
			SalaryRange:     html.UnescapeString(salary),
			JobType:         "Full-time",
			ExperienceLevel: html.UnescapeString(exp),
			Skills:          json.RawMessage(skillsJSON),
			URL:             jobURL,
			PostedAt:        &postedAt,
			ScrapedAt:       time.Now(),
			IsActive:        true,
		})
	}

	fmt.Printf("✅ Extracted %d jobs from Naukri via chromedp\n", len(jobs))
	return jobs, errScrape
}
