package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"shiori/internal/model"
	"time"

	_ "modernc.org/sqlite"
)

const (
	FeedTypeLatest  = "latest"
	FeedTypePopular = "popular"
)

const schema = `
CREATE TABLE IF NOT EXISTS news (
    id TEXT NOT NULL,
    title TEXT NOT NULL,
    url TEXT NOT NULL,
    source TEXT NOT NULL,
    category TEXT,
    published_at DATETIME,
    scraped_at DATETIME NOT NULL,
    feed_type TEXT NOT NULL,
    PRIMARY KEY (url, feed_type)
);

CREATE INDEX IF NOT EXISTS idx_news_source_type ON news(source, feed_type);
CREATE INDEX IF NOT EXISTS idx_news_published ON news(published_at DESC);
`

// OpenDB opens a SQLite database connection and initializes the schema
func OpenDB(path string) (*sql.DB, error) {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create data directory: %w", err)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Enable WAL mode for better concurrent read performance
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("set WAL mode: %w", err)
	}

	// Initialize schema
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("create schema: %w", err)
	}

	return db, nil
}

// SaveNews inserts or updates a news item in the database
func SaveNews(db *sql.DB, news *model.News, feedType string) error {
	query := `
		INSERT INTO news (id, title, url, source, category, published_at, scraped_at, feed_type)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(url, feed_type) DO UPDATE SET
			title = excluded.title,
			category = excluded.category,
			published_at = excluded.published_at,
			scraped_at = excluded.scraped_at
	`

	_, err := db.Exec(query,
		news.ID,
		news.Title,
		news.URL,
		news.Source,
		news.Category,
		news.PublishedAt,
		news.ScrapedAt,
		feedType,
	)
	if err != nil {
		return fmt.Errorf("save news: %w", err)
	}

	return nil
}

// LoadNewsByFeedType loads news grouped by source for a specific feed type
func LoadNewsByFeedType(db *sql.DB, feedType string, limitPerSource int) (map[string][]*model.News, error) {
	// Using a window function to get top N per source
	query := `
		WITH ranked AS (
			SELECT 
				id, title, url, source, category, published_at, scraped_at,
				ROW_NUMBER() OVER (PARTITION BY source ORDER BY published_at DESC) as rn
			FROM news
			WHERE feed_type = ?
		)
		SELECT id, title, url, source, category, published_at, scraped_at
		FROM ranked
		WHERE rn <= ?
		ORDER BY source, published_at DESC
	`

	rows, err := db.Query(query, feedType, limitPerSource)
	if err != nil {
		return nil, fmt.Errorf("query news: %w", err)
	}
	defer rows.Close()

	result := make(map[string][]*model.News)

	for rows.Next() {
		var news model.News
		var publishedAt sql.NullTime

		err := rows.Scan(
			&news.ID,
			&news.Title,
			&news.URL,
			&news.Source,
			&news.Category,
			&publishedAt,
			&news.ScrapedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		if publishedAt.Valid {
			news.PublishedAt = publishedAt.Time
		}

		result[news.Source] = append(result[news.Source], &news)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}

	return result, nil
}

// CleanupOldNews removes news older than the specified duration
func CleanupOldNews(db *sql.DB, maxAge time.Duration) (int64, error) {
	cutoff := time.Now().Add(-maxAge)

	result, err := db.Exec("DELETE FROM news WHERE scraped_at < ?", cutoff)
	if err != nil {
		return 0, fmt.Errorf("cleanup old news: %w", err)
	}

	count, _ := result.RowsAffected()
	return count, nil
}

// GetNewsCount returns total count of news in database
func GetNewsCount(db *sql.DB) (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM news").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count news: %w", err)
	}
	return count, nil
}
