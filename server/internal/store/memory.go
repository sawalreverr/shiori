package store

import (
	"shiori/internal/model"
	"sync"
)

type Store struct {
	news map[string]*model.News
	mu   sync.RWMutex
}

// NewStore creates a new store
func NewStore() *Store {
	return &Store{
		news: make(map[string]*model.News),
	}
}

// Save stores a news, returns true if new
func (s *Store) Save(news *model.News) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.news[news.ID]; exists {
		return false
	}

	s.news[news.ID] = news
	return true
}

// GetAll returns all news
func (s *Store) GetAll() []*model.News {
	s.mu.RLock()
	defer s.mu.RUnlock()

	newsArr := make([]*model.News, 0, len(s.news))
	for _, news := range s.news {
		newsArr = append(newsArr, news)
	}

	return newsArr
}

// GetPopular returns popular news
func (s *Store) GetPopular() []*model.News {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var newsArr []*model.News
	for _, news := range s.news {
		if news.IsPopular {
			newsArr = append(newsArr, news)
		}
	}

	return newsArr
}

// Count returns total news
func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.news)
}
