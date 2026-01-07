package scraper

import (
	"context"
	"fmt"
	"regexp"
	"shiori/internal/model"
	"strings"
	"time"
)

type DetikScraper struct {
	client *HTTPClient
	parser *Parser
}

func NewDetikScraper(client *HTTPClient) *DetikScraper {
	return &DetikScraper{
		client: client,
		parser: NewParser(),
	}
}

// Name returns "detik" for source
func (s *DetikScraper) Name() string {
	return "detik"
}

// ScrapeLatest gets latest news from source
func (s *DetikScraper) ScrapeLatest(ctx context.Context) ([]*model.News, error) {
	url := "https://news.detik.com/indeks"

	body, err := s.client.Fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("fetch detik: %w", err)
	}

	return s.parseNews(string(body))
}

// ScrapePopular gets latest news from source
func (s *DetikScraper) ScrapePopular(ctx context.Context) ([]*model.News, error) {
	url := "https://www.detik.com/terpopuler/news"

	body, err := s.client.Fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("fetch detik: %w", err)
	}

	return s.parseNews(string(body))
}

// parseNews extracts news from HTML
func (s *DetikScraper) parseNews(html string) ([]*model.News, error) {
	var newsArr []*model.News

	pattern := `<div class="media__text">[\s\S]*?</div>`
	re := regexp.MustCompile(pattern)
	matches := re.FindAllString(html, -1)

	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}

		url := s.parser.ExtractBetween(match, `<a href="`, `"`)
		title := s.parser.ExtractBetween(match, `"newsfeed", "`, `"`)
		createdAt := s.parser.ExtractBetween(match, `title="`, `"`)
		createdAt = strings.Replace(strings.SplitN(createdAt, ", ", 2)[1], "WIB", "+0700", 1)
		published, _ := time.Parse("02 Jan 2006 15:04 -0700", createdAt)

		// skip if already seenn and no content
		if url == "" || title == "" || seen[url] {
			continue
		}
		seen[url] = true

		news := &model.News{
			Title:       title,
			URL:         url,
			Source:      s.Name(),
			Category:    "News",
			PublishedAt: published,
			ScrapedAt:   time.Now(),
		}
		news.GenerateID()
		newsArr = append(newsArr, news)
	}

	return newsArr, nil
}
