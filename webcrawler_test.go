package webcrawler

import (
	"encoding/json"
	"os"
	"testing"
)

// Test_CrawlerExample Has been made as a wrapper around the webcrawler package, to demonstrate its functionality.
// Careful with a depth too large, it may cause rate limit blocks on monzo :)
// True is passed into the crawler, to hide duplicates, This has just been done to ensure the json payload is smaller.
func Test_CrawlerExample(t *testing.T) {

	// Initialize our crawler, with sameHost enabled
	crawler, err := New("https://monzo.com/", true, 4)
	if err != nil {
		t.Fatal(err)
	}

	// Start crawling with hide duplicates enabled
	l := crawler.CrawlDepth(20, true)
	if l == nil {
		t.Fatalf("link is nil")
	}

	bytes, err := json.MarshalIndent(l, "", "    ")
	if err != nil {
		t.Fatalf("failed to marshal, err: %v", err)
	}
	saveToFile("monzo_tree_results.json", bytes)

	allLinksBytes, err := json.MarshalIndent(crawler.history.GetKeys(), "", "    ")
	if err != nil {
		t.Fatalf("failed to marshal, err: %v", err)
	}
	saveToFile("monzo_link_list.json", allLinksBytes)
}

func saveToFile(name string, data []byte) {
	f, err := os.Create(name)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	f.Write(data)
}
