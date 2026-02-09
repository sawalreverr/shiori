package scraper

import (
	"context"
	"fmt"
	"shiori/internal/model"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type IDXChannelScraper struct {
	client *HTTPClient
	parser *Parser
}

func NewIDXChannelScraper(client *HTTPClient) *IDXChannelScraper {
	return &IDXChannelScraper{client: client, parser: NewParser()}
}

func (s *IDXChannelScraper) Name() string {
	return "IDXChannel"
}

func (s *IDXChannelScraper) Scrape(ctx context.Context) ([]*model.News, error) {
	var news []*model.News
	start := 9
	seen := make(map[string]bool)

	for i := range 3 {
		url := "https://www.idxchannel.com/indeks"
		if i == 0 {
			url += "?idkanal=1"
		} else {
			url = fmt.Sprintf("%s/more/%d?idkanal=1", url, start)
			start += 6
		}

		body, err := s.client.Fetch(ctx, url)
		if err != nil {
			return nil, fmt.Errorf("fetch idxchannel: %w", err)
		}

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
		if err != nil {
			return nil, fmt.Errorf("parse idxchannel: %w", err)
		}

		doc.Find(".bt-con").Each(func(i int, sel *goquery.Selection) {
			aEl := sel.Find("a")
			link, exists := aEl.Attr("href")
			if !exists {
				return
			}

			if _, ok := seen[link]; ok {
				return
			}
			seen[link] = true

			imgLink, _ := aEl.Find("img").Attr("data-original")
			title := sel.Find(".list-berita-baru").Text()

			publishedEl := sel.Find(".mh-clock").Text()
			published, _ := s.parser.ParseTimeFormat(publishedEl, "02/01/2006 15:04 -0700")

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
