package scraper

import (
	"context"
	"fmt"
	"shiori/internal/model"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type BloombergTechnozScraper struct {
	client *HTTPClient
	parser *Parser
}

func NewBloombergTechnozScraper(client *HTTPClient) *BloombergTechnozScraper {
	return &BloombergTechnozScraper{client: client, parser: NewParser()}
}

func (s *BloombergTechnozScraper) Name() string {
	return "BloombergTechnoz"
}

func (s *BloombergTechnozScraper) Scrape(ctx context.Context) ([]*model.News, error) {
	url := "https://www.bloombergtechnoz.com/indeks/market"
	body, err := s.client.Fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("fetch bloombergtechnoz: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("parse bloombergtechnoz: %w", err)
	}

	var news []*model.News
	seen := make(map[string]bool)

	doc.Find("div.card-box.ft150.margin-bottom-xl").Each(func(i int, sel *goquery.Selection) {
		aEl := sel.Find("a")
		linkHref, exists := aEl.Attr("href")
		if !exists {
			return
		}

		if _, ok := seen[linkHref]; ok {
			return
		}
		seen[linkHref] = true

		title := strings.TrimSpace(sel.Find(".title.margin-bottom-xs").Text())
		imgLink, _ := sel.Find(".img-card img").Attr("src")

		additionsEl := sel.Find(".title.fw4.cl-blue").Text()
		additions := strings.Split(additionsEl, "|")
		if len(additions) == 0 {
			return
		}

		published := s.parser.ParseTime(strings.TrimSpace(additions[1]))

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
