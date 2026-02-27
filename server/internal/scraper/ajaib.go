package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"shiori/internal/model"
	"time"
)

type AjaibResponse struct {
	Posts []AjaibPost `json:"posts"`
}

type AjaibPost struct {
	Slug      string    `json:"slug"`
	Title     string    `json:"title"`
	Date      time.Time `json:"date"`
	Thumbnail string    `json:"thumbnail"`
}

type AjaibScraper struct {
	client *HTTPClient
	parser *Parser
}

func NewAjaibScraper(client *HTTPClient) *AjaibScraper {
	return &AjaibScraper{client: client, parser: NewParser()}
}

func (s *AjaibScraper) Name() string {
	return "Ajaib"
}

func (s *AjaibScraper) Scrape(ctx context.Context) ([]*model.News, error) {
	var news []*model.News

	// ajaib api with 20 post
	url := "https://ajaib.co.id/api/posts-by-categories?slugs=berita&per_page=20&page=1"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("fetch ajaib: %w", err)
	}
	req.Header.Set("User-Agent", s.client.userAgent)

	resp, err := s.client.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch ajaib: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("fetch ajaib: %w", err)
	}

	var posts []AjaibResponse
	if err = json.NewDecoder(resp.Body).Decode(&posts); err != nil {
		return nil, fmt.Errorf("parsing ajaib: %w", err)
	}

	for _, post := range posts[0].Posts {
		news = append(news, &model.News{
			Title:       post.Title,
			URL:         fmt.Sprintf("https://ajaib.co.id/belajar/berita/%s", post.Slug),
			ImageURL:    post.Thumbnail,
			Source:      "ajaib",
			PublishedAt: post.Date,
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
