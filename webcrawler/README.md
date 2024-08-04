# Web Crawler

A simple web crawler implemented in Go that traverses web pages, handles various link types, and maintains a history of visited links. It supports depth-based traversal and manages different types of content and errors.

## Features

- **Depth-based Crawling**: Traverse a web page and its links up to a specified depth.
- **Link Types**: Recognize and categorize different link types including internal, external, hash, and path links.
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

You can use the web crawler by importing it into your Go code.

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

## Development

To contribute to the project:

1. **Fork the Repository** on GitHub.
2. **Create a New Branch**:

    ```sh
    git checkout -b feature/new-feature
    ```

3. **Commit Your Changes**:

    ```sh
    git add .
    git commit -m "Add new feature"
    ```

4. **Push to Your Fork**:

    ```sh
    git push origin feature/new-feature
    ```

5. **Open a Pull Request** on GitHub.
