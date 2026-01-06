package scraper

import (
	"context"
	"fmt"
	"regexp"
	"shiori/internal/model"
	"strings"
	"time"
)

type KompasScraper struct {
	client *HTTPClient
	parser *Parser
}

func NewKompasScraper(client *HTTPClient) *KompasScraper {
	return &KompasScraper{
		client: client,
		parser: NewParser(),
	}
}

// Name returns "kompas" for source
func (s *KompasScraper) Name() string {
	return "kompas"
}

// ScrapeLatest gets latest news from source
func (s *KompasScraper) ScrapeLatest(ctx context.Context) ([]*model.News, error) {
	url := "https://indeks.kompas.com/?site=news"

	body, err := s.client.Fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("fetch kompas: %w", err)
	}

	return s.parseNews(string(body), false)
}

// ScrapePopular gets popular news from source
func (s *KompasScraper) ScrapePopular(ctx context.Context) ([]*model.News, error) {
	url := "https://indeks.kompas.com/terpopuler?site=news"

	body, err := s.client.Fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("fetch kompas: %w", err)
	}

	return s.parseNews(string(body), true)
}

// parseNews extracts news from HTML
func (s *KompasScraper) parseNews(html string, isPopular bool) ([]*model.News, error) {
	var newsArr []*model.News

	pattern := `<a[^>]+href="https://[a-z]+\.kompas\.com/read/[^"]+"[\s\S]*?</a>`
	re := regexp.MustCompile(pattern)
	matches := re.FindAllString(html, -1)

	seen := make(map[string]bool) // no dup

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}

		url := s.parser.ExtractBetween(match, `href="`, `"`)
		title := s.parser.ExtractBetween(match, `<h2 class="articleTitle">`, `</h2>`)
		category := s.parser.ExtractBetween(match, `articlePost-subtitle ">`, `</div>`)

		title = s.parser.StripTags(title)
		title = strings.TrimSpace(title)
		category = strings.TrimSpace(category)

		// skip if already seen and no content
		if url == "" || title == "" || seen[url] {
			continue
		}
		seen[url] = true

		news := &model.News{
			Title:       title,
			URL:         url,
			Source:      s.Name(),
			Category:    category,
			IsPopular:   isPopular,
			PublishedAt: time.Now(),
			ScrapedAt:   time.Now(),
		}
		news.GenerateID()
		newsArr = append(newsArr, news)
	}
	return newsArr, nil
}
