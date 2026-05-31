package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

func fetchUrl(url string) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error for %s: %v\n", url, err)
		return
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	content := string(body)
	if len(content) > 200 {
		content = content[:200]
	}
	
	fmt.Printf("URL: %s\nStatus: %d\nContent: %s...\n\n", url, resp.StatusCode, strings.ReplaceAll(content, "\n", ""))
}

func main() {
	fetchUrl("https://www.naukri.com/software-engineer-jobs")
	fetchUrl("https://www.hirist.tech/search/software-engineer-jobs.html")
}
