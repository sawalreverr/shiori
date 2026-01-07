package model

import (
	"crypto/md5"
	"encoding/hex"
	"time"
)

type News struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	ImageURL    string    `json:"image_url,omitempty"`
	Source      string    `json:"source"`
	Category    string    `json:"category,omitempty"`
	Author      string    `json:"author,omitempty"`
	PublishedAt time.Time `json:"published_at"`
	ScrapedAt   time.Time `json:"scraped_at"`
}

func (a *News) GenerateID() {
	hash := md5.Sum([]byte(a.URL))
	a.ID = hex.EncodeToString(hash[:])
}
