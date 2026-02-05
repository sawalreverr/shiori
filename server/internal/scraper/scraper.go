package scraper

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"shiori/internal/config"
	"shiori/internal/model"
	"sync"
	"time"
)

type Scraper interface {
	Name() string
	Scrape(ctx context.Context) ([]*model.News, error)
}

type HTTPClient struct {
	client    *http.Client
	userAgent string
	delay     time.Duration
}

func NewHTTPClient(cfg *config.Config) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
		userAgent: cfg.UserAgent,
		delay:     cfg.Delay,
	}
}

// Fetch returns HTML as bytes
func (c *HTTPClient) Fetch(ctx context.Context, url string) ([]byte, error) {
	time.Sleep(c.delay)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("bad status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	return body, nil
}

type Manager struct {
	scrapers []Scraper
	client   *HTTPClient
}

func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		scrapers: []Scraper{},
		client:   NewHTTPClient(cfg),
	}
}

// Register adds a scraper
func (m *Manager) Register(s Scraper) {
	m.scrapers = append(m.scrapers, s)
}

// GetHTTPClient returns the HTTP client
func (m *Manager) GetHTTPClient() *HTTPClient {
	return m.client
}

func (m *Manager) ScrapeAll(ctx context.Context) ([]*model.News, []error) {
	var (
		allNews   []*model.News
		allErrors []error
		mu        sync.Mutex
		wg        sync.WaitGroup
	)

	for _, s := range m.scrapers {
		wg.Add(1)
		go func(scraper Scraper) {
			defer wg.Done()

			scrapeCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			news, err := scraper.Scrape(scrapeCtx)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				allErrors = append(allErrors, fmt.Errorf("%s: %w", scraper.Name(), err))
			} else {
				allNews = append(allNews, news...)
			}
		}(s)
	}

	wg.Wait()
	return allNews, allErrors
}
