package webcrawler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
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

	allLinksBytes, err := crawler.GetHistory()
	if err != nil {
		t.Fatalf("failed to get history, err: %v", err)
	}
	saveToFile("monzo_link_list.json", allLinksBytes)
}

// saveToFile writes the given data to a file.
func saveToFile(filename string, data []byte) error {
	return os.WriteFile(filename, data, 0644)
}

// Helper function to create a mock server with given handlers.
func createMockServer(handlerFunc http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(handlerFunc))
}

// TestResolveLinkType tests the link type resolver.
func TestResolveLinkType(t *testing.T) {
	tests := []struct {
		href     string
		expected string
	}{
		{"http://example.com", PageLink},
		{"https://example.com", PageLink},
		{"#section", HashLink},
		{"/path", PathLink},
		{"unknown", UnknownLink},
	}

	for _, tt := range tests {
		result := resolveLinkType(tt.href)
		if result != tt.expected {
			t.Errorf("resolveLinkType(%q) = %q; want %q", tt.href, result, tt.expected)
		}
	}
}

// TestCrawlDepth tests the crawling functionality with specified depth.
func TestCrawlDepth(t *testing.T) {
	// Mock server for depth 1
	mockServerDepth1 := createMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
			<html>
				<body>
					<a href="/page1">Page 1</a>
					<a href="/page2">Page 2</a>
					<a href="/page3">Page 3</a>
				</body>
			</html>
		`))
	})

	defer func() {
		mockServerDepth1.Close()
	}()

	tests := []struct {
		url                 string
		depth               int
		hideDuplicates      bool
		expectedChildren    []string
		expectedSecondLevel []string
	}{
		{mockServerDepth1.URL, 1, true, []string{"/page1", "/page2", "/page3"}, nil},
		{mockServerDepth1.URL, 2, false, []string{"/page1", "/page2", "/page3"}, []string{"/page1", "/page2", "/page3"}},
	}

	for _, tt := range tests {
		crawler, err := New(tt.url, true, 5)
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}

		rootLink := crawler.CrawlDepth(tt.depth, tt.hideDuplicates)

		// Check first level children
		if len(rootLink.Children) != len(tt.expectedChildren) {
			t.Fatalf("Expected %d first-level children, got %d", len(tt.expectedChildren), len(rootLink.Children))
		}
		for i, expectedHref := range tt.expectedChildren {
			if rootLink.Children[i].Href != expectedHref {
				t.Fatalf("Expected first-level link href to be %v, got %v", expectedHref, rootLink.Children[i].Href)
			}
		}

		// Check second level children if depth > 1
		if tt.depth > 1 {
			for _, child := range rootLink.Children {
				if len(child.Children) != len(tt.expectedSecondLevel) {
					t.Fatalf("Expected %d second-level children for %v, got %d", len(tt.expectedSecondLevel), child.Href, len(child.Children))
				}
				for i, expectedHref := range tt.expectedSecondLevel {
					if child.Children[i].Href != expectedHref {
						t.Fatalf("Expected second-level link href to be %v, got %v", expectedHref, child.Children[i].Href)
					}
				}
			}
		}
	}
}

// TestCrawlHandlesNonHTMLContent tests handling of non-HTML content types.
func TestCrawlHandlesNonHTMLContent(t *testing.T) {
	mockServer := createMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"key":"value"}`))
	})
	defer mockServer.Close()

	crawler, err := New(mockServer.URL, true, 5)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	rootLink := crawler.Crawl()
	if rootLink.Err == nil {
		t.Fatal("Expected error due to non-HTML content type, got none")
	}
}

// TestCrawlHandlesErrorStatusCodes tests handling of non-200 status codes.
func TestCrawlHandlesErrorStatusCodes(t *testing.T) {
	mockServer := createMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	defer mockServer.Close()

	crawler, err := New(mockServer.URL, true, 5)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	rootLink := crawler.Crawl()
	if rootLink.Err == nil {
		t.Fatal("Expected error due to non-200 status code, got none")
	}
}

// TestCrawlHandlesEmptyResponse tests handling of empty HTTP responses.
func TestCrawlHandlesEmptyResponse(t *testing.T) {
	mockServer := createMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/html")
	})
	defer mockServer.Close()

	crawler, err := New(mockServer.URL, true, 5)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	rootLink := crawler.Crawl()
	if rootLink.Err == nil {
		t.Fatal("Expected error due to empty response, got none")
	}
}

// TestGetHistory tests that the history returns the correct data.
func TestGetHistory(t *testing.T) {
	mockServer := createMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
			<html>
				<body>
					<a href="/page1">Page 1</a>
				</body>
			</html>
		`))
	})
	defer mockServer.Close()

	crawler, err := New(mockServer.URL, true, 5)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	crawler.Crawl()

	historyData, err := crawler.GetHistory()
	if err != nil {
		t.Fatalf("GetHistory() error = %v", err)
	}

	var links []string
	if err := json.Unmarshal(historyData, &links); err != nil {
		t.Fatalf("Failed to unmarshal history data: %v", err)
	}

	if len(links) != 2 {
		t.Fatal("Expected history to contain at least one link, got none")
	}

}

// TestCrawlHandlesTimeout tests handling of request timeouts.
func TestCrawlHandlesTimeout(t *testing.T) {
	mockServer := createMockServer(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // Simulate a timeout
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/html")
	})
	defer mockServer.Close()

	crawler, err := New(mockServer.URL, true, 1) // Set timeout to 1 second
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	rootLink := crawler.Crawl()
	if rootLink.Err == nil {
		t.Fatal("Expected error due to request timeout, got none")
	}
}
