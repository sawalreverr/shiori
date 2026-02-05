package model

import (
	"crypto/md5"
	"encoding/hex"
	"time"
)

type News struct {
	ID          string
	Title       string
	URL         string
	ImageURL    string
	Source      string
	PublishedAt time.Time
	CreatedAt   time.Time
}

type MarketNewsResponse struct {
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	ImageURL    string    `json:"image_url,omitzero"`
	PublishedAt time.Time `json:"published_at,omitzero"`
}

type SourceGroupResponse struct {
	Source string               `json:"id"`
	News   []MarketNewsResponse `json:"news"`
}

type MarketResponse struct {
	Status        string                `json:"status"`
	SourceCount   int                   `json:"source_count"`
	TotalNews     int                   `json:"total_news"`
	LastScrapedAt time.Time             `json:"last_scraped_at,omitzero"`
	Items         []SourceGroupResponse `json:"items"`
}

func (a *News) GenerateID() {
	hash := md5.Sum([]byte(a.URL))
	a.ID = hex.EncodeToString(hash[:])
}
