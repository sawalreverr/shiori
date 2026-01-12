package scraper

import (
	"context"
	"fmt"
	"regexp"
	"shiori/internal/model"
	"time"
)

type TribunNewsScraper struct {
	client *HTTPClient
	parser *Parser
}

func NewTribunNewsScraper(client *HTTPClient) *TribunNewsScraper {
	return &TribunNewsScraper{
		client: client,
		parser: NewParser(),
	}
}

// Name returns "tribunnews" for source
func (s *TribunNewsScraper) Name() string {
	return "tribunnews"
}

// ScrapeLatest get latest news from source
func (s *TribunNewsScraper) ScrapeLatest(ctx context.Context) ([]*model.News, error) {
	url := "https://www.tribunnews.com/index-news/news?date="

	body, err := s.client.Fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("fetch tribunnews: %w", err)
	}

	return s.parseNews(string(body))
}

// ScrapePopular get popular news from source
func (s *TribunNewsScraper) ScrapePopular(ctx context.Context) ([]*model.News, error) {
	url := "https://www.tribunnews.com/populer/?section=&type=12h"

	body, err := s.client.Fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("fetch tribunnews: %w", err)
	}

	return s.parsePopular(string(body))
}

// parseNews extracts news from HTML
func (s *TribunNewsScraper) parseNews(html string) ([]*model.News, error) {
	var newsArr []*model.News

	pattern := `<li[^"]+class="ptb15">[\s\S]*?</li>`
	re := regexp.MustCompile(pattern)
	matches := re.FindAllString(html, -1)

	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}

		aHref := s.parser.ExtractBetween(match, `<h3 class="f16 fbo">`, ` title`)
		url := s.parser.ExtractBetween(aHref, `<a href="`, `"`)
		title := s.parser.ExtractBetween(match, fmt.Sprintf(`"%s" title="`, url), `"`)
		published_at := s.parser.ExtractBetween(match, `<time class="grey">`, `</time>`)
		published, _ := s.parser.ParseTimeFormat(published_at, "02 January 2006 15:04 -0700")

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

// parsePopular extracts news from HTML
func (s *TribunNewsScraper) parsePopular(html string) ([]*model.News, error) {
	var newsArr []*model.News

	pattern := `<li[^"]+class="pt5 pb20">[\s\S]*?</li>`
	re := regexp.MustCompile(pattern)
	matches := re.FindAllString(html, -1)

	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}

		url := s.parser.ExtractBetween(match, `<a href="`, `"`)
		title := s.parser.ExtractBetween(match, fmt.Sprintf(`"%s" title="`, url), `"`)
		published_at := s.parser.ExtractBetween(match, `<time class="grey pt5">`, `</time>`)
		published, _ := s.parser.ParseTimeFormat(published_at, "02 January 2006 15:04 -0700")

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
