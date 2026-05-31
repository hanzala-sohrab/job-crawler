package main

import (
	"fmt"
	"strings"

	"github.com/gocolly/colly/v2"
)

func main() {
	fmt.Println("Testing Naukri...")
	c1 := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
	)
	c1.OnHTML("script", func(e *colly.HTMLElement) {
		if strings.Contains(e.Text, "jobDetails") || strings.Contains(e.Text, "__NEXT_DATA__") || strings.Contains(e.Text, "initialState") {
			fmt.Println("Naukri: Found potential JSON data in script tag")
		}
	})
	c1.OnHTML(".jobTuple", func(e *colly.HTMLElement) {
		fmt.Println("Naukri: Found jobTuple element")
	})
	c1.OnHTML(".srp-jobtuple-wrapper", func(e *colly.HTMLElement) {
		fmt.Println("Naukri: Found srp-jobtuple-wrapper element")
	})
	c1.Visit("https://www.naukri.com/software-engineer-jobs")


	fmt.Println("\nTesting Hirist...")
	c2 := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
	)
	c2.OnHTML("script", func(e *colly.HTMLElement) {
		if strings.Contains(e.Text, "__NEXT_DATA__") {
			fmt.Println("Hirist: Found __NEXT_DATA__")
		}
	})
	c2.OnHTML(".job-card", func(e *colly.HTMLElement) {
		fmt.Println("Hirist: Found job-card element")
	})
	c2.Visit("https://www.hirist.tech/search/software-engineer-jobs.html")
}
