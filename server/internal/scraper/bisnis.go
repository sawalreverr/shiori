package scraper

import (
	"context"
	"fmt"
	"shiori/internal/model"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type BisnisIndonesiaScraper struct {
	client *HTTPClient
	parser *Parser
}

func NewBisnisIndonesiaScraper(client *HTTPClient) *BisnisIndonesiaScraper {
	return &BisnisIndonesiaScraper{client: client, parser: NewParser()}
}

func (s *BisnisIndonesiaScraper) Name() string {
	return "Bisnis Indonesia"
}

func (s *BisnisIndonesiaScraper) Scrape(ctx context.Context) ([]*model.News, error) {
	url := "https://www.bisnis.com/index?categoryId=194&date=&type=indeks"
	body, err := s.client.Fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("fetch bisnis: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("parse bisnis: %w", err)
	}

	var news []*model.News
	seen := make(map[string]bool)

	doc.Find("div#indeksListView div.artWrap div.artItem").Each(func(i int, sel *goquery.Selection) {
		aEl := sel.Find("a")
		linkHref, exists := aEl.Attr("href")
		if !exists {
			return
		}

		if _, ok := seen[linkHref]; ok {
			return
		}
		seen[linkHref] = true

		titleEl := sel.Find("h4.artTitle")
		title := strings.TrimSpace(titleEl.Text())

		imgEl := sel.Find("img")
		imgSrc, exists := imgEl.Attr("src")
		if !exists {
			return
		}

		publishedEl := sel.Find("div.artDate")
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

	for _, n := range news {
		n.GenerateID()
	}

	if len(news) == 0 {
		return news, fmt.Errorf("no news found")
	}

	return news, nil
}
