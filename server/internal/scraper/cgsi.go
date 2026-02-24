package scraper

import (
	"context"
	"fmt"
	"shiori/internal/model"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type CGSIScraper struct {
	client *HTTPClient
	parser *Parser
}

func NewCGSIScraper(client *HTTPClient) *CGSIScraper {
	return &CGSIScraper{client: client, parser: NewParser()}
}

func (s *CGSIScraper) Name() string {
	return "CGSI"
}

func (s *CGSIScraper) Scrape(ctx context.Context) ([]*model.News, error) {
	var news []*model.News
	seen := make(map[string]bool)
	pageTotal := 2

	for page := range pageTotal {
		url := fmt.Sprintf("https://itrade.cgsi.co.id/blog?category=1&page=%d", page+1)

		body, err := s.client.Fetch(ctx, url)
		if err != nil {
			return nil, fmt.Errorf("fetch cgsi: %w", err)
		}

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
		if err != nil {
			return nil, fmt.Errorf("parse cgsi: %w", err)
		}

		doc.Find(".column.column-33").Each(func(i int, sel *goquery.Selection) {
			aEl := sel.Find("a").First()
			link, exists := aEl.Attr("href")
			if !exists {
				return
			}

			if _, ok := seen[link]; ok {
				return
			}
			seen[link] = true

			title := strings.TrimSpace(sel.Find(".title-list").Text())

			imgLink, _ := sel.Find("img").First().Attr("src")

			publishedAt := sel.Find(".date").Text()
			published, _ := s.parser.ParseTimeFormat(publishedAt, "2 Jan 2006")

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
