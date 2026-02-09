package scraper

import (
	"context"
	"fmt"
	"shiori/internal/model"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type KabarbursaScraper struct {
	client *HTTPClient
	parser *Parser
}

func NewKabarbursaScraper(client *HTTPClient) *KabarbursaScraper {
	return &KabarbursaScraper{client: client, parser: NewParser()}
}

func (s *KabarbursaScraper) Name() string {
	return "Kabarbursa"
}

func (s *KabarbursaScraper) Scrape(ctx context.Context) ([]*model.News, error) {
	var news []*model.News
	seen := make(map[string]bool)
	pageTotal := 2

	for page := range pageTotal {
		url := fmt.Sprintf("https://www.kabarbursa.com/market-hari-ini?page=%d", page+1)

		body, err := s.client.Fetch(ctx, url)
		if err != nil {
			return nil, fmt.Errorf("fetch kabarbursa: %w", err)
		}

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
		if err != nil {
			return nil, fmt.Errorf("parse kabarbursa: %w", err)
		}

		doc.Find("article").Each(func(i int, sel *goquery.Selection) {
			aEl := sel.Find("a")
			link, exists := aEl.Attr("href")
			if !exists {
				return
			}

			if _, ok := seen[link]; ok {
				return
			}
			seen[link] = true

			imgLink, _ := sel.Find("img").Attr("src")
			title := strings.TrimSpace(aEl.Text())

			publishedEl := strings.TrimSpace(sel.Find("span").Text())
			publishedTxt := strings.Split(publishedEl, "•")[1]
			published, _ := s.parser.ParseTimeFormat(publishedTxt, "02 January 2006")

			news = append(news, &model.News{
				Title:       title,
				URL:         link,
				ImageURL:    imgLink,
				Source:      s.Name(),
				PublishedAt: published,
			})
		})
	}

	for _, n := range news {
		n.GenerateID()
	}

	if len(news) == 0 {
		return news, fmt.Errorf("no news found")
	}

	return news, nil
}
