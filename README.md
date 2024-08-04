# Web Crawler

A simple web crawler implemented in Go that traverses web pages, handles various link types, and maintains a history of visited links. It supports depth-based traversal and manages different types of content and errors.

## Features

- **Depth-based Crawling**: Traverse a web page and its links up to a specified depth.
- **Link Types**: Recognize and categorize different link types including internal, external, hash, and path links.
- **Tree View**: Supports generating a tree view of the web page link structure in json format.
- **History Management**: Maintain a history of visited links and provide a snapshot of the history in JSON format.
- **Error Handling**: Manage various errors such as non-HTML content, status codes, and timeouts.

## Installation

1. **Navigate to the Project Directory**:

```sh
cd webcrawler
```

2. **Install Dependencies**:

```sh
go mod tidy
```

## Usage

You can use the web crawler by importing it into your Go code or by using the command-line interface provided.

### Command-Line Usage

Build the command-line tool:

```sh
go build crawler.go
```

Run the crawler:

```sh
./crawler -url <URL> -depth <DEPTH> -hide-duplicates -output-file <OUTPUT_FILE>
```

- `-url`: The URL to start crawling from.
- `-same-domain`: Whether to only crawl within the same domain
- `-timeout`: "HTTP client timeout in seconds
- `-depth`: The depth of traversal (default is 2).
- `-hide-duplicates`: Hide duplicate links during traversal.
- `-output-file`: File to save the crawled links
- `-history-file`: File to save the history of links

#### Example

To crawl `https://monzo.com/"` up to depth 3, hiding duplicates, and save the tree results to `results.json`, and links list to `results_links.json`:

```sh
./crawler -url "https://monzo.com/" -same-domain=true -timeout=5 -depth=3 -hide-duplicates=true -output-file="results.json" -history-file="results_links.json"
```

## Testing

Run the unit tests to ensure the functionality of the web crawler:

```sh
go test ./... -cover -v
```

### Test Results

Here’s a summary of the test results:

```
=== RUN   Test_CrawlerExample
--- PASS: Test_CrawlerExample (2.66s)
=== RUN   TestResolveLinkType
--- PASS: TestResolveLinkType (0.00s)
=== RUN   TestCrawlDepth
--- PASS: TestCrawlDepth (0.00s)
=== RUN   TestCrawlHandlesNonHTMLContent
--- PASS: TestCrawlHandlesNonHTMLContent (0.00s)
=== RUN   TestCrawlHandlesErrorStatusCodes
--- PASS: TestCrawlHandlesErrorStatusCodes (0.00s)
=== RUN   TestCrawlHandlesEmptyResponse
--- PASS: TestCrawlHandlesEmptyResponse (0.00s)
=== RUN   TestGetHistory
--- PASS: TestGetHistory (0.00s)
=== RUN   TestCrawlHandlesTimeout
--- PASS: TestCrawlHandlesTimeout (2.01s)
PASS
coverage: 95.1% of statements
ok      webcrawler/webcrawler   4.882s  coverage: 95.1% of statements
```