package store

import (
	"database/sql"
	"log"
	"shiori/internal/model"
	"sort"
	"sync"
)

const (
	MaxNewsPerSource = 100
	DefaultLimit     = 20
)

type Store struct {
	bySource map[string][]*model.News // source -> all articles
	count    int
	mu       sync.RWMutex
	db       *sql.DB
	feedType string // "latest" or "popular"
}

// NewStore creates a new store (in-memory only, for backward compatibility)
func NewStore() *Store {
	return &Store{
		bySource: make(map[string][]*model.News),
	}
}

// NewStoreWithDB creates a store backed by SQLite
func NewStoreWithDB(db *sql.DB, feedType string) *Store {
	return &Store{
		bySource: make(map[string][]*model.News),
		db:       db,
		feedType: feedType,
	}
}

type SourceGroup struct {
	Source string
	News   []*model.News
}

// LoadFromDB populates the in-memory cache from SQLite
func (s *Store) LoadFromDB() error {
	if s.db == nil {
		return nil // No DB configured, skip
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := LoadNewsByFeedType(s.db, s.feedType, MaxNewsPerSource)
	if err != nil {
		return err
	}

	s.bySource = data
	s.count = 0
	for _, news := range data {
		s.count += len(news)
	}

	log.Printf("Loaded %d news items for %s feed from database", s.count, s.feedType)
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
		if err := SaveNews(s.db, news, s.feedType); err != nil {
			log.Printf("Failed to save to database: %v", err)
			// Continue anyway to keep in-memory working
		}
	}

	// add to source list
	s.bySource[source] = append([]*model.News{news}, s.bySource[source]...)
	s.count++

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

	return groups
}

// Count returns total news
func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.count
}
