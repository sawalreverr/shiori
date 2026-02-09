package scraper

import (
	"context"
	"fmt"
	"regexp"
	"shiori/internal/model"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type IDNFinansialsScraper struct {
	client *HTTPClient
	parser *Parser
}

func NewIDNFinansialsScraper(client *HTTPClient) *IDNFinansialsScraper {
	return &IDNFinansialsScraper{client: client, parser: NewParser()}
}

func (s *IDNFinansialsScraper) Name() string {
	return "IDNFinansials"
}

func (s *IDNFinansialsScraper) Scrape(ctx context.Context) ([]*model.News, error) {
	var news []*model.News
	seen := make(map[string]bool)
	pageTotal := 2

	for page := range pageTotal {
		url := fmt.Sprintf("https://www.idnfinancials.com/id/news/page/%d", page+1)

		body, err := s.client.Fetch(ctx, url)
		if err != nil {
			return nil, fmt.Errorf("fetch idnfinancials: %w", err)
		}

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
		if err != nil {
			return nil, fmt.Errorf("parse idnfinancials: %w", err)
		}

		// featured news
		doc.Find(".featured-news .first").Each(func(i int, sel *goquery.Selection) {
			aEl := sel.Find("a").First()
			link, exists := aEl.Attr("href")
			if !exists {
				return
			}

			if _, ok := seen[link]; ok {
				return
			}
			seen[link] = true

			title := sel.Find("h1.title").Text()

			imgStyle, _ := sel.Find(".st-image").Attr("style")
			imgLink := extractURLImg(imgStyle)

			publishedDt := sel.Find(".date-published").AttrOr("data-date", "")
			published, _ := time.Parse(time.RFC3339, publishedDt)

			news = append(news, &model.News{
				Title:       title,
				URL:         link,
				ImageURL:    imgLink,
				Source:      s.Name(),
				PublishedAt: published,
			})
		})

		// grid news
		doc.Find(".featured-news .ln-item").Each(func(i int, sel *goquery.Selection) {
			aEl := sel.Find("a").First()
			link, exists := aEl.Attr("href")
			if !exists {
				return
			}

			if _, ok := seen[link]; ok {
				return
			}
			seen[link] = true

			title := sel.Find("h1.title").Text()

			imgStyle, _ := sel.Find(".image").Attr("style")
			imgLink := extractURLImg(imgStyle)

			publishedDt := sel.Find(".date-published").AttrOr("data-date", "")
			published, _ := time.Parse(time.RFC3339, publishedDt)

			news = append(news, &model.News{
				Title:       title,
				URL:         link,
				ImageURL:    imgLink,
				Source:      s.Name(),
				PublishedAt: published,
			})
		})

		// latest news
		doc.Find(".widget.latest-news article.item").Each(func(i int, sel *goquery.Selection) {
			aEl := sel.Find("h2.title a").First()
			link, exists := aEl.Attr("href")
			if !exists {
				return
			}

			if _, ok := seen[link]; ok {
				return
			}
			seen[link] = true

			title := strings.TrimSpace(aEl.Text())

			imgStyle, _ := sel.Find(".image").Attr("style")
			imgLink := extractURLImg(imgStyle)

			publishedDt := sel.Find(".date-published").AttrOr("data-date", "")
			published, _ := time.Parse(time.RFC3339, publishedDt)

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

func extractURLImg(style string) string {
	re := regexp.MustCompile(`background-image:\s*url\((.*?)\);`)
	match := re.FindStringSubmatch(style)
	if len(match) > 1 {
		return match[1]
	}

	return ""
}
