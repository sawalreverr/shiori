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

const schema = `
CREATE TABLE IF NOT EXISTS news (
    id TEXT NOT NULL,
    title TEXT NOT NULL,
    url TEXT NOT NULL,
    image_url TEXT,
    source TEXT NOT NULL,
    published_at DATETIME,
    created_at DATETIME NOT NULL,
    PRIMARY KEY (url)
);

CREATE INDEX IF NOT EXISTS idx_news_source ON news(source);
CREATE INDEX IF NOT EXISTS idx_news_published ON news(published_at DESC);
`

// OpenDB opens a SQLite database connection and initializes the schema
func OpenDB(path string) (*sql.DB, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create data directory: %w", err)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("set WAL mode: %w", err)
	}

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("create schema: %w", err)
	}

	return db, nil
}

// SaveNews inserts or updates a news item in database
func SaveNews(db *sql.DB, news *model.News) error {
	query := `
		INSERT INTO news (id, title, url, image_url, source, published_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(url) DO UPDATE SET
			title = excluded.title,
			image_url = excluded.image_url,
			published_at = excluded.published_at,
			created_at = excluded.created_at
	`

	_, err := db.Exec(query,
		news.ID,
		news.Title,
		news.URL,
		news.ImageURL,
		news.Source,
		news.PublishedAt,
		news.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("save news: %w", err)
	}

	return nil
}

func LoadNews(db *sql.DB, limitPerSource int) (map[string][]*model.News, error) {
	query := `
		WITH ranked AS (
			SELECT
				id, title, url, image_url, source, published_at, created_at,
				ROW_NUMBER() OVER (PARTITION BY source ORDER BY published_at DESC) as rn
			FROM news
		)
		SELECT id, title, url, image_url, source, published_at, created_at
		FROM ranked
		WHERE rn <= ?
		ORDER BY source, published_at DESC
	`

	rows, err := db.Query(query, limitPerSource)
	if err != nil {
		return nil, fmt.Errorf("query news: %w", err)
	}
	defer rows.Close()

	result := make(map[string][]*model.News)

	for rows.Next() {
		var news model.News
		var publishedAt sql.NullTime
		var imageURL sql.NullString
		var createdAt sql.NullTime

		err := rows.Scan(
			&news.ID,
			&news.Title,
			&news.URL,
			&imageURL,
			&news.Source,
			&publishedAt,
			&createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		if publishedAt.Valid {
			news.PublishedAt = publishedAt.Time
		}

		if imageURL.Valid {
			news.ImageURL = imageURL.String
		}

		if createdAt.Valid {
			news.CreatedAt = createdAt.Time
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

	result, err := db.Exec("DELETE FROM news WHERE created_at < ?", cutoff)
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
