package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func main() {
	url := "https://www.naukri.com/jobapi/v3/search?noOfResults=20&urlType=search_by_keyword&searchType=adv&keyword=software%20engineer"

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	req.Header.Set("appid", "109")
	req.Header.Set("systemid", "109")
	req.Header.Set("clientid", "d3s")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Status: %d\n", resp.StatusCode)
	
	body, _ := io.ReadAll(resp.Body)
	
	var data map[string]interface{}
	json.Unmarshal(body, &data)
	
	if jobs, ok := data["jobDetails"].([]interface{}); ok {
		fmt.Printf("Found %d jobs\n", len(jobs))
		if len(jobs) > 0 {
			firstJob := jobs[0].(map[string]interface{})
			fmt.Printf("First job title: %v\n", firstJob["title"])
		}
	} else {
		fmt.Println("No jobDetails array found in response")
		fmt.Printf("Response preview: %s\n", string(body))
	}
}
