package webcrawler

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"

	"webcrawler/history"
)

const (
	defaultDepth = 2
	maxDepth     = 20
)

// LINK types
const (
	HashLink     = "hashlink"     // typically represents a position on the page to link to.
	PathLink     = "pathlink"     // represents a path on the same domain as its parent
	PageLink     = "pagelink"     // represents a link to a different domain than its parent
	UnknownLink  = "unknownlink"  // represents an unknown link type
	ExistingLink = "existinglink" // represents a link that has already been visited
)

// Webcrawler provides functionality for crawling web pages
type Webcrawler struct {
	href     string
	url      *url.URL
	sameHost bool
	timeout  time.Duration
	history  *history.History[*Link]
}

// Link represents a linked list node, showing the tree of links crawled over
type Link struct {
	Parent   *Link   `json:"-"`
	Text     string  `json:"text"`
	Href     string  `json:"href"`
	Err      error   `json:"err,omitempty"`
	LinkType string  `json:"linkType"`
	Children []*Link `json:"children,omitempty"`
}

// New creates a new Webcrawler
func New(href string, sameDomain bool, timeout time.Duration) (*Webcrawler, error) {
	// Parse the given URL
	url, err := url.Parse(href)
	if err != nil {
		return nil, err
	}

	// Initialize and return a Webcrawler instance
	return &Webcrawler{
		href:     href,
		url:      url,
		sameHost: sameDomain,
		timeout:  timeout,
		history:  history.NewHistory[*Link](),
	}, nil
}

// Crawl will traverse throughout the given URL using the default depth
func (c *Webcrawler) Crawl() *Link {
	return c.CrawlDepth(defaultDepth, false)
}

// CrawlDepth will traverse throughout the given URL with specified depth and duplicate settings
func (c *Webcrawler) CrawlDepth(depth int, hideDuplicates bool) *Link {
	// Set depth to defaultDepth if it's zero or less
	if depth <= 0 {
		depth = defaultDepth
	} else if depth > maxDepth {
		depth = maxDepth
	}

	// Start traversing from the root link
	return c.traverse(&Link{Href: c.href}, depth, hideDuplicates)
}

// traverse handles visiting the link and processing its children
func (c *Webcrawler) traverse(l *Link, depth int, hideDuplicates bool) *Link {
	// Visit the current link and find all links on the page
	l.visit(c.timeout)
	// Record the link in the history
	c.history.Add(l.Href, l)

	var wg sync.WaitGroup

	for i := 0; i < len(l.Children); i++ {
		child := l.Children[i]

		// Check if we've visited this link before
		if _, ok := c.history.Check(child.Href); ok {
			if hideDuplicates {
				// Hide duplicate links
				l.Children = append(l.Children[:i], l.Children[i+1:]...)
				i--
				continue
			}
			child.LinkType = ExistingLink
			continue
		}

		// Add the child link to history
		c.history.Add(child.Href, child)

		// Skip hash links
		if child.LinkType == HashLink {
			continue
		}

		// Enforce same host policy if applicable
		if c.sameHost && child.LinkType != PathLink {
			childURL, err := url.Parse(child.Href)
			if err != nil {
				child.Err = err
				continue
			}
			if c.url.Host != childURL.Host {
				continue
			}
		}

		// Traverse children if depth is greater than 0
		if depth > 0 {
			wg.Add(1)
			go func(ch *Link) {
				defer wg.Done()
				c.traverse(ch, depth-1, hideDuplicates)
			}(child)
		}
	}

	// Wait for all goroutines to finish
	wg.Wait()

	return l
}

// visit retrieves the page and extracts links from it
func (l *Link) visit(timeout time.Duration) *Link {
	url := l.Href

	// Add host to path link if needed
	if l.LinkType == PathLink && !strings.HasPrefix(l.Href, l.Parent.Href) {
		url = l.Parent.Href + l.Href
	}

	resp, err := http.Get(url)
	if err != nil {
		l.Err = err
		return l
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		l.Err = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		return l
	}

	if !strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
		l.Err = fmt.Errorf("unexpected content type: %v", resp.Header.Get("Content-Type"))
		return l
	}

	return l.parseLinks(resp.Body)
}

// parseLinks extracts links from the HTTP response body
func (l *Link) parseLinks(body io.ReadCloser) *Link {

	t := html.NewTokenizer(body)
	l.Children = make([]*Link, 0, 40)
	depth := 0
	var currentLink *Link

	for {
		tt := t.Next()
		switch tt {
		case html.ErrorToken:
			l.Err = t.Err()
			return l
		case html.TextToken:
			if depth > 0 && currentLink != nil {
				currentLink.Text = string(t.Text())
				l.Children = append(l.Children, currentLink)
				currentLink = nil
			}
		case html.StartTagToken:
			tn, hasAttr := t.TagName()
			if string(tn) == "a" && hasAttr {
				depth++
				currentLink = &Link{Parent: l}
				for {
					key, val, moreAttr := t.TagAttr()
					if string(key) == "href" {
						currentLink.Href = string(val)
						currentLink.LinkType = resolveLinkType(currentLink.Href)
					}
					if !moreAttr {
						break
					}
				}
			}
		case html.EndTagToken:
			tn, _ := t.TagName()
			if string(tn) == "a" {
				depth--
			}
		}
	}
}

// resolveLinkType returns the link type based on the href
func resolveLinkType(href string) string {
	switch {
	case strings.HasPrefix(href, "http://"), strings.HasPrefix(href, "https://"):
		return PageLink
	case strings.HasPrefix(href, "#"):
		return HashLink
	case strings.HasPrefix(href, "/"):
		return PathLink
	default:
		return UnknownLink
	}
}
