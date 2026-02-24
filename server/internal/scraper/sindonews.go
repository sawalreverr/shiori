package scraper

import (
	"context"
	"fmt"
	"shiori/internal/model"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type SindonewsScraper struct {
	client *HTTPClient
	parser *Parser
}

func NewSindonewsScraper(client *HTTPClient) *SindonewsScraper {
	return &SindonewsScraper{client: client, parser: NewParser()}
}

func (s *SindonewsScraper) Name() string {
	return "sindonews"
}

func (s *SindonewsScraper) Scrape(ctx context.Context) ([]*model.News, error) {
	var news []*model.News
	seen := make(map[string]bool)

	url := "https://www.sindonews.com/indeks/8"

	body, err := s.client.Fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("fetch sindonews: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("parse sindonews: %w", err)
	}

	doc.Find(".list-article").Each(func(i int, sel *goquery.Selection) {
		aEl := sel.Find(".title-article a").First()
		link, exists := aEl.Attr("href")
		if !exists {
			return
		}

		if _, ok := seen[link]; ok {
			return
		}
		seen[link] = true

		title := aEl.Text()

		var imgLink string
		_, exists = sel.Find(".img-article img").Attr("loading")
		if exists {
			imgLink, _ = sel.Find(".img-article img").Attr("data-src")
		} else {
			imgLink, _ = sel.Find(".img-article img").Attr("src")
		}

		publishedTxt := sel.Find(".date-article").Text()
		publishedNoDay := strings.Split(publishedTxt, ",")[1]
		published, _ := s.parser.ParseTimeFormat(publishedNoDay, "2 January 2006 - 15:04 -0700")

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
