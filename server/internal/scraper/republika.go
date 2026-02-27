package scraper

import (
	"context"
	"fmt"
	"shiori/internal/model"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type RepublikaScraper struct {
	client *HTTPClient
	parser *Parser
}

func NewRepublikaScraper(client *HTTPClient) *RepublikaScraper {
	return &RepublikaScraper{client: client, parser: NewParser()}
}

func (s *RepublikaScraper) Name() string {
	return "Republika"
}

func (s *RepublikaScraper) Scrape(ctx context.Context) ([]*model.News, error) {
	var news []*model.News
	seen := make(map[string]bool)

	url := "https://ekonomi.republika.co.id/"

	body, err := s.client.Fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("fetch republika: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("parse republika: %w", err)
	}

	doc.Find(".conten1").Each(func(i int, sel *goquery.Selection) {
		aEl := sel.Find("a").First()
		link, exists := aEl.Attr("href")
		if !exists {
			return
		}

		if _, ok := seen[link]; ok {
			return
		}
		seen[link] = true

		title := aEl.Find("h3").Text()

		imgLink, _ := aEl.Find("img").Attr("data-original")

		publishedAt := strings.Split(aEl.Find(".date").Nodes[0].LastChild.Data, "- ")[1]

		var published time.Time
		if strings.Contains(publishedAt, "yang lalu") {
			published = s.parser.ParseTime(publishedAt)
		} else {
			published, _ = s.parser.ParseTimeFormat(publishedAt, "02 January 2006 15:04")
		}

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
