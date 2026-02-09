package scraper

import (
	"context"
	"fmt"
	"shiori/internal/model"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type KatadataScraper struct {
	client *HTTPClient
	parser *Parser
}

func NewKatadataScraper(client *HTTPClient) *KatadataScraper {
	return &KatadataScraper{client: client, parser: NewParser()}
}

func (s *KatadataScraper) Name() string {
	return "Katadata"
}

func (s *KatadataScraper) Scrape(ctx context.Context) ([]*model.News, error) {
	url := "https://katadata.co.id/finansial"

	body, err := s.client.Fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("fetch katadata: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("parse katadata: %w", err)
	}

	var news []*model.News
	seen := make(map[string]bool)

	// headline slider news
	doc.Find(".container article").Each(func(i int, sel *goquery.Selection) {
		aEl := sel.Find("a")
		linkHref, exists := aEl.Attr("href")
		if !exists {
			return
		}

		if _, ok := seen[linkHref]; ok {
			return
		}
		seen[linkHref] = true

		title := strings.TrimSpace(sel.Find("h2").Text())

		body, err := s.client.Fetch(ctx, linkHref)
		if err != nil {
			return
		}

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
		if err != nil {
			return
		}

		imgLink, _ := doc.Find(".detail-image img.img-fullwidth").Attr("src")

		published_at := doc.Find(".detail-date").Text()
		published, _ := s.parser.ParseTimeFormat(published_at, "2 January 2006 15:04")

		news = append(news, &model.News{
			Title:       title,
			URL:         linkHref,
			ImageURL:    imgLink,
			Source:      s.Name(),
			PublishedAt: published,
		})
	})

	// headline items news
	doc.Find(".headline-item").Each(func(i int, sel *goquery.Selection) {
		aEl := sel.Find("a")
		linkHref, exists := aEl.Attr("href")
		if !exists {
			return
		}

		if _, ok := seen[linkHref]; ok {
			return
		}
		seen[linkHref] = true

		title := strings.TrimSpace(sel.Find("h2").Text())

		body, err := s.client.Fetch(ctx, linkHref)
		if err != nil {
			return
		}

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
		if err != nil {
			return
		}

		imgLink, _ := doc.Find(".detail-image img.img-fullwidth").Attr("src")

		published_at := doc.Find(".detail-date").Text()
		published, _ := s.parser.ParseTimeFormat(published_at, "2 January 2006 15:04")

		news = append(news, &model.News{
			Title:       title,
			URL:         linkHref,
			ImageURL:    imgLink,
			Source:      s.Name(),
			PublishedAt: published,
		})
	})

	doc.Find("article.article.article--berita.d-flex").Each(func(i int, sel *goquery.Selection) {
		aEl := sel.Find("a")
		linkHref, exists := aEl.Attr("href")
		if !exists {
			return
		}

		if _, ok := seen[linkHref]; ok {
			return
		}
		seen[linkHref] = true

		title := strings.TrimSpace(sel.Find("h3").Text())

		imgLink, _ := sel.Find(".content-image.scale img").Attr("data-src")

		published_at := sel.Find(".article__date").Text() // • 6 Februari 2026, 15.27
		published_at = strings.Split(published_at, "•")[1]
		published, _ := s.parser.ParseTimeFormat(published_at, "2 January 2006 15.04")

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
