package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"webcrawler/webcrawler"
)

// saveToFile writes the given data to a file.
func saveToFile(filename string, data []byte) error {
	return os.WriteFile(filename, data, 0644)
}

// This crawler client is a simple cli wrapper for the webcrawler package
// Example command:
// go run crawler.go -url "https://monzo.com/" -same-domain=true -timeout=5 -depth=20 -hide-duplicates=true -output-file="monzo_tree_results.json" -history-file="monzo_link_list.json"
func main() {
	startURL := flag.String("url", "https://example.com", "The start URL for the crawler")
	sameDomain := flag.Bool("same-domain", true, "Whether to only crawl within the same domain")
	timeout := flag.Int("timeout", 5, "HTTP client timeout in seconds")
	depth := flag.Int("depth", 2, "Depth of crawling")
	hideDuplicates := flag.Bool("hide-duplicates", false, "Whether to hide duplicate links")
	outputFile := flag.String("output-file", "", "File to save the crawled links")
	historyFile := flag.String("history-file", "", "File to save the history of links")

	flag.Parse()

	crawler, err := webcrawler.New(*startURL, *sameDomain, *timeout)
	if err != nil {
		log.Fatalf("Failed to create crawler: %v", err)
	}

	rootLink := crawler.CrawlDepth(*depth, *hideDuplicates)

	if *outputFile != "" {
		result, err := json.MarshalIndent(rootLink, "", "  ")
		if err != nil {
			log.Fatalf("Failed to marshal result: %v", err)
		}
		if err := saveToFile(*outputFile, result); err != nil {
			log.Fatalf("Failed to save result to file: %v", err)
		}
	}

	history, err := crawler.GetHistory()
	if *historyFile != "" {
		if err != nil {
			log.Fatalf("Failed to get history: %v", err)
		}
		if err := saveToFile(*historyFile, history); err != nil {
			log.Fatalf("Failed to save history to file: %v", err)
		}
	}

	fmt.Println(string(history))
}
