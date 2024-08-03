package webcrawler

import (
	"encoding/json"
	"os"
	"testing"
)

func TestCrawler(t *testing.T) {

	crawler := New("https://monzo.com/")

	l := crawler.Crawl()
	if l == nil {
		t.Fatalf("link is nil")
	}

	bytes, err := json.Marshal(l)
	if err != nil {
		t.Fatalf("failed to marshal")
	}

	f, err := os.Create("monzo_results.json")
	if err != nil {
		t.Fatalf("failed to marshal")
	}
	defer f.Close()
	f.Write(bytes)
}
