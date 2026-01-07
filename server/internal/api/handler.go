package api

import (
	"encoding/json"
	"net/http"
	"shiori/internal/store"
	"time"
)

type Handler struct {
	store *store.Store
}

func NewHandler(s *store.Store) *Handler {
	return &Handler{store: s}
}

// RegisterRoutes sets up all routes
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.Health)
	mux.HandleFunc("/api/news", h.GetNews)
	mux.HandleFunc("/api/news/popular", h.GetPopular)
}

// Health returns server status
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// GetNews returns all news
func (h *Handler) GetNews(w http.ResponseWriter, r *http.Request) {
	news := h.store.GetAll()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status": "success",
		"items":  news,
		"count":  len(news),
	})
}

// GetPopular returns popular news
func (h *Handler) GetPopular(w http.ResponseWriter, r *http.Request) {
	news := h.store.GetPopular()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status": "success",
		"items":  news,
		"count":  len(news),
	})
}
