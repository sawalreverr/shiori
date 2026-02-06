package scraper

import (
	"context"
	"fmt"
	"shiori/internal/model"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type StockWatchScraper struct {
	client *HTTPClient
	parser *Parser
}

func NewStockWatchScraper(client *HTTPClient) *StockWatchScraper {
	return &StockWatchScraper{client: client, parser: NewParser()}
}

func (s *StockWatchScraper) Name() string {
	return "StockWatch"
}

func (s *StockWatchScraper) Scrape(ctx context.Context) ([]*model.News, error) {
	var news []*model.News
	seen := make(map[string]bool)
	totalPage := 2

	for page := 1; page < totalPage+1; page++ {
		url := fmt.Sprintf("https://stockwatch.id/category/market/page/%d/", page)
		body, err := s.client.Fetch(ctx, url)
		if err != nil {
			return nil, fmt.Errorf("fetch stockwatch: %w", err)
		}

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
		if err != nil {
			return nil, fmt.Errorf("parse stockwatch: %w", err)
		}

		doc.Find(".vc_column.tdi_85 .td-cpt-post").Each(func(i int, sel *goquery.Selection) {
			aEl := sel.Find(".td-module-title a")
			linkHref, exists := aEl.Attr("href")
			if !exists {
				return
			}

			if _, ok := seen[linkHref]; ok {
				return
			}
			seen[linkHref] = true

			title := aEl.Text()

			imgEl := sel.Find(".entry-thumb.td-thumb-css")
			imgSrc, exists := imgEl.Attr("data-img-url")
			if !exists {
				return
			}

			publishedEl := sel.Find(".td-module-date")
			publishedDt, exists := publishedEl.Attr("datetime")
			if !exists {
				return
			}
			published, _ := time.Parse(time.RFC3339, publishedDt)

			news = append(news, &model.News{
				Title:       title,
				URL:         linkHref,
				ImageURL:    imgSrc,
				Source:      s.Name(),
				PublishedAt: published,
			})
		})

		doc.Find(".vc_column.tdi_95 .td-cpt-post").Each(func(i int, sel *goquery.Selection) {
			aEl := sel.Find(".td-module-title a")
			linkHref, exists := aEl.Attr("href")
			if !exists {
				return
			}

			if _, ok := seen[linkHref]; ok {
				return
			}
			seen[linkHref] = true

			title := aEl.Text()

			imgEl := sel.Find(".entry-thumb.td-thumb-css")
			imgSrc, exists := imgEl.Attr("data-img-url")
			if !exists {
				return
			}

			publishedEl := sel.Find(".td-module-date")
			publishedDt, exists := publishedEl.Attr("datetime")
			if !exists {
				return
			}
			published, _ := time.Parse(time.RFC3339, publishedDt)

			news = append(news, &model.News{
				Title:       title,
				URL:         linkHref,
				ImageURL:    imgSrc,
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
