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
	Category    string
	Author      string
	PublishedAt time.Time
	ScrapedAt   time.Time
}

type NewsResponse struct {
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	Category    string    `json:"category"`
	PublishedAt time.Time `json:"published_at,omitzero"`
}

type SourceGroupResponse struct {
	Source string         `json:"id"`
	News   []NewsResponse `json:"news"`
}

func (a *News) GenerateID() {
	hash := md5.Sum([]byte(a.URL))
	a.ID = hex.EncodeToString(hash[:])
}
