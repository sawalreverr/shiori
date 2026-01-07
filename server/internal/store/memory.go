package store

import (
	"shiori/internal/model"
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
}

// NewStore creates a new store
func NewStore() *Store {
	return &Store{
		bySource: make(map[string][]*model.News),
	}
}

type SourceGroup struct {
	Source string        `json:"id"`
	News   []*model.News `json:"news"`
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

	// add to source list
	s.bySource[source] = append([]*model.News{news}, s.bySource[source]...)
	s.count++

	// trim if too many
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
