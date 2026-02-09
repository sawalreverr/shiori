package scraper

import (
	"context"
	"fmt"
	"shiori/internal/model"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type DetikFinanceScraper struct {
	client *HTTPClient
	parser *Parser
}

func NewDetikFinanceScraper(client *HTTPClient) *DetikFinanceScraper {
	return &DetikFinanceScraper{client: client, parser: NewParser()}
}

func (s *DetikFinanceScraper) Name() string {
	return "Detik Finance"
}

func (s *DetikFinanceScraper) Scrape(ctx context.Context) ([]*model.News, error) {
	url := "https://finance.detik.com/indeks"
	body, err := s.client.Fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("fetch detikfinance: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("parse detikfinance: %w", err)
	}

	var news []*model.News
	seen := make(map[string]bool)

	doc.Find("article").Each(func(i int, sel *goquery.Selection) {
		aEl := sel.Find("a")
		linkHref, exists := aEl.Attr("href")
		if !exists {
			return
		}

		if _, ok := seen[linkHref]; ok {
			return
		}
		seen[linkHref] = true

		title := strings.TrimSpace(sel.Find("h3.media__title").Text())
		imgLink, _ := sel.Find("img").Attr("src")

		publishedEl := sel.Find("div.media__date")
		published := s.parser.ParseTime(publishedEl.Text())

		news = append(news, &model.News{
			Title:       title,
			URL:         linkHref,
			ImageURL:    imgLink,
			Source:      s.Name(),
			PublishedAt: published,
		})
	})

	for _, n := range news {
		n.GenerateID()
	}

	if len(news) == 0 {
		return news, fmt.Errorf("no news found")
	}

	return news, nil
}
