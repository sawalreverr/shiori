package scraper

import (
	"context"
	"fmt"
	"regexp"
	"shiori/internal/model"
	"time"
)

type CNNIndonesiaScraper struct {
	client *HTTPClient
	parser *Parser
}

func NewCNNIndonesiaScraper(client *HTTPClient) *CNNIndonesiaScraper {
	return &CNNIndonesiaScraper{
		client: client,
		parser: NewParser(),
	}
}

// Name returns "cnnindonesia" for source
func (s *CNNIndonesiaScraper) Name() string {
	return "cnnindonesia"
}

// ScrapeLatest get latest news from source
func (s *CNNIndonesiaScraper) ScrapeLatest(ctx context.Context) ([]*model.News, error) {
	url := "https://www.cnnindonesia.com/indeks"

	body, err := s.client.Fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("fetch cnnindonesia: %w", err)
	}

	return s.parseNews(string(body))
}

// ScrapePopular get popular news from source
func (s *CNNIndonesiaScraper) ScrapePopular(ctx context.Context) ([]*model.News, error) {
	url := "https://www.cnnindonesia.com/"

	body, err := s.client.Fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("fetch cnnindonesia: %w", err)
	}

	return s.parsePopular(string(body))
}

// parseNews extracts news from HTML
func (s *CNNIndonesiaScraper) parseNews(html string) ([]*model.News, error) {
	var newsArr []*model.News

	pattern := `<article[^>]class="flex-grow">[\s\S]*?</article>`
	re := regexp.MustCompile(pattern)
	matches := re.FindAllString(html, -1)

	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}

		url := s.parser.ExtractBetween(match, `href="`, `"`)
		title := s.parser.ExtractBetween(match, `text-cnn_red">`, `</h2>`)
		category := s.parser.ExtractBetween(match, `text-xs text-cnn_red">`, `</span>`)
		published_at := s.parser.ExtractBetween(match, `text-xs text-cnn_black_light3"> • `, `<`)
		published := s.parser.ParseTime(published_at)

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

// parsePopular extracts news from HTML
func (s *CNNIndonesiaScraper) parsePopular(html string) ([]*model.News, error) {
	var newsArr []*model.News

	pattern := `<article[^>]class="pl-9 mb-4 relative">[\s\S]*?</article>`
	re := regexp.MustCompile(pattern)
	matches := re.FindAllString(html, -1)

	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}

		url := s.parser.ExtractBetween(match, `href="`, `"`)
		title := s.parser.ExtractBetween(match, `text-cnn_red">`, `</h2>`)
		category := s.parser.ExtractBetween(match, `text-xs text-cnn_red">`, `</span>`)

		if url == "" || title == "" || seen[url] {
			continue
		}
		seen[url] = true

		news := &model.News{
			Title:     title,
			URL:       url,
			Source:    s.Name(),
			Category:  category,
			ScrapedAt: time.Now(),
		}
		news.GenerateID()
		newsArr = append(newsArr, news)
	}

	return newsArr, nil
}
