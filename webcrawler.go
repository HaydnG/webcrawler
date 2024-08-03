package webcrawler

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"golang.org/x/net/html"

	"webcrawler/history"
)

const (
	defaultDepth = 2
	maxDepth     = 20
)

// LINK types
const (
	// typically represents a position on the page to link to.
	HashLink = "hashlink"
	// represents a path on the same domain as its parent
	PathLink = "pathlink"
	// represents a link to a different domain than its parent
	PageLink = "pagelink"

	UnkownLink = "unkownlink"

	ExistingLink = "existinglink"
)

// Webcrawler provides functionality for crawling web pages
type Webcrawler struct {
	url string

	WebHits      int
	LinksCreated int

	// a dictionary of already visited links, avoid visiting them twice
	histoy *history.History[*link]
}

type link struct {
	Parent   *link  `json:"-"`
	Text     string `json:"text"`
	Href     string `json:"href"`
	Err      error  `json:"err,omitempty"`
	LinkType string `json:"linkType"`

	Children []*link `json:"children,omitempty"`
}

// New creates a new Webcrawler
func New(url string) *Webcrawler {
	return &Webcrawler{
		url:    url,
		histoy: history.NewHistory[*link](),
	}
}

func (c *Webcrawler) Crawl() *link {
	return c.CrawlDepth(defaultDepth)
}

func (c *Webcrawler) CrawlDepth(depth int) *link {
	if depth == 0 {
		depth = defaultDepth
	}

	if depth > maxDepth {
		depth = maxDepth
	}

	l := c.traverse(&link{Href: c.url}, depth)

	return l
}

func (c *Webcrawler) traverse(l *link, depth int) *link {

	if depth == 0 {
		return l
	}

	// find all links on the page
	l.visit()
	// make a record in the history
	c.histoy.Add(l.Href, l)

	wg := &sync.WaitGroup{}

	for i := 0; i < len(l.Children); i++ {
		if l.Children[i].LinkType == HashLink {
			// Hash links, just link to different positions on the same page
			continue
		}

		if l.Children[i].Href == l.Href {
			continue // Dont want to visit the same page twice
		}

		// Check if we've visited this link before
		_, ok := c.histoy.Check(l.Children[i].Href)
		if ok {
			l.Children[i].LinkType = ExistingLink
			continue
		}

		wg.Add(1)
		go func(child *link) {
			c.traverse(child, depth-1)
			wg.Done()
		}(l.Children[i])
	}

	wg.Wait()

	return l
}

// visit takes a given link, and returns all links on that page
func (l *link) visit() *link {
	url := l.Href

	// a Path link is on the same host, so we need to check if we need to add the host back on
	if l.LinkType == PathLink && !strings.Contains(l.Href, l.Parent.Href) {
		url = l.Parent.Href + l.Href
	}

	resp, err := http.Get(url)
	if err != nil {
		l.Err = err
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		l.Err = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		return nil
	}

	if !strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
		l.Err = fmt.Errorf("unexpected content type, %v", resp.Header.Get("Content-Type"))
		return nil
	}

	return l.parseLinks(resp.Body)
}

func (l *link) parseLinks(body io.ReadCloser) *link {
	t := html.NewTokenizer(body)

	l.Children = make([]*link, 0, 40)
	depth := 0
	var currentLink *link
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
			if len(tn) == 1 && tn[0] == 'a' {
				depth++
				if hasAttr {
					currentLink = &link{Parent: l}
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
			}
		case html.EndTagToken:
			tn, _ := t.TagName()
			if len(tn) == 1 && tn[0] == 'a' {
				depth--
			}
		}
	}
}

func resolveLinkType(href string) string {
	switch {
	case strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://"):
		return PageLink
	case strings.HasPrefix(href, "#"):
		return HashLink
	case strings.HasPrefix(href, "/"):
		return PathLink
	}
	return UnkownLink
}
