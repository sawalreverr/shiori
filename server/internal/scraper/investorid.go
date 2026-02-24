package scraper

import (
	"context"
	"fmt"
	"shiori/internal/model"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type InvestorIDScraper struct {
	client *HTTPClient
	parser *Parser
}

func NewInvestorIDScraper(client *HTTPClient) *InvestorIDScraper {
	return &InvestorIDScraper{client: client, parser: NewParser()}
}

func (s *InvestorIDScraper) Name() string {
	return "InvestorID"
}

func (s *InvestorIDScraper) Scrape(ctx context.Context) ([]*model.News, error) {
	var news []*model.News
	seen := make(map[string]bool)

	url := "https://investor.id/market/indeks"

	body, err := s.client.Fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("fetch investor id: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("parse investor id: %w", err)
	}

	doc.Find(".row.mb-4.position-relative").Each(func(i int, sel *goquery.Selection) {
		aEl := sel.Find("a").First()
		link, exists := aEl.Attr("href")
		if !exists {
			return
		}

		fullURL := fmt.Sprintf("https://investor.id%s", link)

		if _, ok := seen[link]; ok {
			return
		}
		seen[link] = true

		title := sel.Find("h4").Text()

		imgLink, _ := aEl.Find("img").Attr("src")

		publishedAt := strings.TrimSpace(sel.Find(".text-muted.small").Text())
		published := s.parser.ParseTime(publishedAt)

		news = append(news, &model.News{
			Title:       title,
			URL:         fullURL,
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
