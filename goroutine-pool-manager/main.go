package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"runtime"
	"time"

	"github.com/gammazero/workerpool"
)

type ScrapeResult struct {
	URL      string
	Title    string
	Keywords []string
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	urls := []string{
		"https://golang.org",
		"https://google.com",
		"https://github.com",
	}

	numWorkers := min(len(urls), runtime.NumCPU())
	pool := workerpool.New(numWorkers)

	resultsCh := make(chan ScrapeResult, len(urls))

	for _, url := range urls {
		url := url
		pool.Submit(func() {
			result := scrape(ctx, url)
			if result != nil {
				resultsCh <- *result
			}
		})
	}

	pool.StopWait()
	close(resultsCh)

	for result := range resultsCh {
		fmt.Printf("Scraped URL %s: Title = %s, Keywords = %v\n", result.URL, result.Title, result.Keywords)
	}

	fmt.Println("Scrape completed.")
}

func scrape(ctx context.Context, url string) *ScrapeResult {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		fmt.Printf("Error creating request for URL %s: %v\n", url, err)
		return nil
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Error fetching URL %s: %v\n", url, err)
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body for URL %s: %v\n", url, err)
		return nil
	}

	titleRe := regexp.MustCompile(`<title>(.*?)</title>`)
	keywordsRe := regexp.MustCompile(`<meta name="keywords" content="(.*?)" ?/?>`)

	titleMatch := titleRe.FindStringSubmatch(string(body))
	keywordsMatch := keywordsRe.FindStringSubmatch(string(body))

	var title string
	if len(titleMatch) > 1 {
		title = titleMatch[1]
	} else {
		title = "N/A"
	}

	var keywords []string
	if len(keywordsMatch) > 1 {
		keywords = regexp.MustCompile(`,\s*`).Split(keywordsMatch[1], -1)
	}

	return &ScrapeResult{
		URL:      url,
		Title:    title,
		Keywords: keywords,
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
