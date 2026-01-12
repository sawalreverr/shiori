package scraper

import (
	"context"
	"fmt"
	"regexp"
	"shiori/internal/model"
	"time"
)

type Liputan6Scraper struct {
	client *HTTPClient
	parser *Parser
}

func NewLiputan6Scraper(client *HTTPClient) *Liputan6Scraper {
	return &Liputan6Scraper{
		client: client,
		parser: NewParser(),
	}
}

// Name returns "liputan6" for source
func (s *Liputan6Scraper) Name() string {
	return "liputan6"
}

// ScrapeLatest get latest news from source
func (s *Liputan6Scraper) ScrapeLatest(ctx context.Context) ([]*model.News, error) {
	url := "https://www.liputan6.com/news/indeks"

	body, err := s.client.Fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("fetch liputan6: %w", err)
	}

	return s.parseNews(string(body))
}

// ScrapePopular get popular news from source
func (s *Liputan6Scraper) ScrapePopular(ctx context.Context) ([]*model.News, error) {
	url := "https://www.liputan6.com/news/indeks/terpopuler"

	body, err := s.client.Fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("fetch liputan6: %w", err)
	}

	return s.parseNews(string(body))
}

// parseNews extracts news from HTML
func (s *Liputan6Scraper) parseNews(html string) ([]*model.News, error) {
	var newsArr []*model.News

	pattern := `<aside[^>]+class="articles--rows--item__details">[\s\S]*?</aside>`
	re := regexp.MustCompile(pattern)
	matches := re.FindAllString(html, -1)

	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}

		url := s.parser.ExtractBetween(match, `<h4 class="articles--rows--item__title"><a href="`, `"`)
		title := s.parser.ExtractBetween(match, `data-template-var="title">`, `</span>`)
		category := s.parser.ExtractBetween(match, `data-template-var="category">`, `</a>`)
		published_at := s.parser.ExtractBetween(match, `datetime="`, `"`)

		published, _ := time.Parse(time.RFC3339, published_at)

		if url == "" || title == "" || seen[url] {
			continue
		}
		seen[url] = true

		news := &model.News{
			Title:       title,
			URL:         url,
			Source:      s.Name(),
			Category:    category,
			PublishedAt: published,
			ScrapedAt:   time.Now(),
		}
		news.GenerateID()
		newsArr = append(newsArr, news)
	}

	return newsArr, nil
}
