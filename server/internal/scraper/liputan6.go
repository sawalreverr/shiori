package scraper

import (
	"context"
	"fmt"
	"shiori/internal/model"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Liputan6Scraper struct {
	client *HTTPClient
	parser *Parser
}

func NewLiputan6Scraper(client *HTTPClient) *Liputan6Scraper {
	return &Liputan6Scraper{client: client, parser: NewParser()}
}

func (s *Liputan6Scraper) Name() string {
	return "Liputan6"
}

func (s *Liputan6Scraper) Scrape(ctx context.Context) ([]*model.News, error) {
	var news []*model.News
	seen := make(map[string]bool)

	url := "https://www.liputan6.com/bisnis/indeks"

	body, err := s.client.Fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("fetch liputan6: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("parse liputan6: %w", err)
	}

	doc.Find(".articles--rows--item").Each(func(i int, sel *goquery.Selection) {
		aEl := sel.Find("a").First()
		link, exists := aEl.Attr("href")
		if !exists {
			return
		}

		if _, ok := seen[link]; ok {
			return
		}
		seen[link] = true

		title := sel.Find("h4").Text()

		imgLink, _ := aEl.Find("img").Attr("src")

		publishedDt, _ := sel.Find("time").Attr("datetime")
		published, _ := time.Parse(time.RFC3339, publishedDt)

		news = append(news, &model.News{
			Title:       title,
			URL:         link,
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
