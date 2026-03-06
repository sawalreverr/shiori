package store

import (
	"database/sql"
	"log/slog"
	"shiori/internal/model"
	"sort"
	"sync"
	"time"
)

const (
	MaxNewsPerSource = 100
	DefaultLimit     = 20
)

type Store struct {
	bySource      map[string][]*model.News // source -> all articles
	count         int
	lastScrapedAt time.Time
	mu            sync.RWMutex
	db            *sql.DB
}

// NewStore creates a new store (in-memory only, for backward compatibility)
func NewStore() *Store {
	return &Store{
		bySource: make(map[string][]*model.News),
	}
}

// NewStoreWithDB creates a store backed by SQLite
func NewStoreWithDB(db *sql.DB) *Store {
	return &Store{
		bySource: make(map[string][]*model.News),
		db:       db,
	}
}

type SourceGroup struct {
	Source string
	News   []*model.News
}

// LoadFromDB populates the in-memory cache from SQLite
func (s *Store) LoadFromDB() error {
	if s.db == nil {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := LoadNews(s.db, MaxNewsPerSource)
	if err != nil {
		return err
	}

	s.bySource = data
	s.count = 0
	for _, news := range data {
		s.count += len(news)
	}

	slog.Info("Loaded news items from database", "count", s.count)
	return nil
}

// Save stores a news, returns true if new
func (s *Store) Save(news *model.News) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	source := news.Source

	// check duplicate
	for _, existing := range s.bySource[source] {
		if existing.ID == news.ID {
			return false
		}
	}

	// Persist to database first (if configured)
	if s.db != nil {
		if err := SaveNews(s.db, news); err != nil {
			slog.Error("Failed to save to database", "error", err)
		}
	}

	// add to source list
	s.bySource[source] = append([]*model.News{news}, s.bySource[source]...)
	s.count++
	s.lastScrapedAt = time.Now()

	// sort by published_at (newest first)
	sort.Slice(s.bySource[source], func(i, j int) bool {
		return s.bySource[source][i].PublishedAt.After(s.bySource[source][j].PublishedAt)
	})

	// trim if too many (remove oldest)
	if len(s.bySource[source]) > MaxNewsPerSource {
		s.bySource[source] = s.bySource[source][:MaxNewsPerSource]
		s.count--
	}

	return true
}

// GetGrouped returns news grouped by source with limit
func (s *Store) GetGrouped(limit int) []SourceGroup {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 {
		limit = DefaultLimit
	}

	groups := make([]SourceGroup, 0, len(s.bySource))
	for source, news := range s.bySource {
		count := len(news)
		if count > limit {
			news = news[:limit]
		}
		groups = append(groups, SourceGroup{source, news})
	}

	// Sort groups alphabetically by source name
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Source < groups[j].Source
	})

	return groups
}

// Count returns total news
func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.count
}

func (s *Store) GetLastScrapedAt() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastScrapedAt
}

// GetSourceCount returns the number of sources
func (s *Store) GetSourceCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.bySource)
}
