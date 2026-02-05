package scraper

import (
	"context"
	"fmt"
	"shiori/internal/model"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type CNBCIndonesiaScraper struct {
	client *HTTPClient
	parser *Parser
}

func NewCNBCIndonesiaScraper(client *HTTPClient) *CNBCIndonesiaScraper {
	return &CNBCIndonesiaScraper{client: client, parser: NewParser()}
}

func (s *CNBCIndonesiaScraper) Name() string {
	return "CNBC Indonesia"
}

func (s *CNBCIndonesiaScraper) Scrape(ctx context.Context) ([]*model.News, error) {
	var news []*model.News
	seen := make(map[string]bool)
	totalPage := 2 // get until page 2 for more news, 10 news each page

	for page := 1; page < totalPage+1; page++ {
		url := fmt.Sprintf("https://www.cnbcindonesia.com/market/indeks/5?page=%d", page)

		body, err := s.client.Fetch(ctx, url)
		if err != nil {
			return nil, fmt.Errorf("fetch cnbc: %w", err)
		}

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
		if err != nil {
			return nil, fmt.Errorf("parse cnbc: %w", err)
		}

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

			titleEl := sel.Find("h2")
			title := strings.TrimSpace(titleEl.Text())

			imgEl := sel.Find("img")
			imgSrc, exists := imgEl.Attr("src")
			if !exists {
				return
			}

			publishedEl := sel.Find(".text-xs.text-gray")
			published := s.parser.ParseTime(publishedEl.Text())

			news = append(news, &model.News{
				Title:       title,
				URL:         linkHref,
				ImageURL:    imgSrc,
				Source:      s.Name(),
				PublishedAt: published,
				CreatedAt:   time.Now(),
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
