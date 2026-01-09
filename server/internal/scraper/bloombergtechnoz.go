package scraper

import (
	"context"
	"fmt"
	"regexp"
	"shiori/internal/model"
	"time"
)

type BloombergTechnozScraper struct {
	client *HTTPClient
	parser *Parser
}

func NewBloomberTechnozScraper(client *HTTPClient) *BloombergTechnozScraper {
	return &BloombergTechnozScraper{
		client: client,
		parser: NewParser(),
	}
}

// Name returns "bloombergtechnoz" for source
func (s *BloombergTechnozScraper) Name() string {
	return "bloombergtechnoz"
}

// ScrapeLatest get latest news from source
func (s *BloombergTechnozScraper) ScrapeLatest(ctx context.Context) ([]*model.News, error) {
	url := "https://www.bloombergtechnoz.com/indeks"

	body, err := s.client.Fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("fetch bloombergtechnoz: %w", err)
	}

	return s.parseNews(string(body))
}

// ScrapePopular get popular news from source
func (s *BloombergTechnozScraper) ScrapePopular(ctx context.Context) ([]*model.News, error) {
	url := "https://www.bloombergtechnoz.com/"

	body, err := s.client.Fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("fetch bloombergtechnoz %w", err)
	}

	return s.parsePopular(string(body))
}

// parseNews extracts news from HTML
func (s *BloombergTechnozScraper) parseNews(html string) ([]*model.News, error) {
	var newsArr []*model.News

	pattern := `<div class="card-box ft150 margin-bottom-xl">[\s\S]*?</a>`
	re := regexp.MustCompile(pattern)
	matches := re.FindAllString(html, -1)

	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}

		url := s.parser.ExtractBetween(match, `<a href="`, `">`)
		title := s.parser.ExtractBetween(match, `<h2 class="title margin-bottom-xs">`, `</h2>`)
		category := s.parser.ExtractBetween(match, `<h6 class="title fw4 cl-blue">`, `<`)
		published_at := s.parser.ExtractBetween(match, `<span class="cl-gray">|`, `</span>`)

		category = s.parser.DecodeHTMLEntities(category)
		published := s.parser.ParseTime(published_at)

		// skip if already seen or no content
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

// parsePopular extract popular
func (s *BloombergTechnozScraper) parsePopular(html string) ([]*model.News, error) {
	var newsArr []*model.News

	ulPattern := regexp.MustCompile(`<ul[^>]*class="[^"]*row-list list-terpopuler[^"]*"[^>]*>[\s\S]*?</ul>`)
	ulMatch := ulPattern.FindString(html)

	liPattern := regexp.MustCompile(`<li[^>]*>[\s\S]*?</li>`)
	liMatch := liPattern.FindAllString(ulMatch, -1)

	seen := make(map[string]bool)

	for _, match := range liMatch {
		if len(match) < 3 {
			continue
		}

		url := s.parser.ExtractBetween(match, `<a href="`, `">`)
		title := s.parser.ExtractBetween(match, `<h5 class="title">`, `</h5>`)

		// skip if already seen or no content
		if url == "" || title == "" || seen[url] {
			continue
		}
		seen[url] = true

		news := &model.News{
			Title:     title,
			URL:       url,
			Source:    s.Name(),
			Category:  "News",
			ScrapedAt: time.Now(),
		}
		news.GenerateID()
		newsArr = append(newsArr, news)
	}

	return newsArr, nil
}
