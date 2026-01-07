package api

import (
	"encoding/json"
	"net/http"
	"shiori/internal/store"
	"strconv"
	"time"
)

type Handler struct {
	latestStore  *store.Store
	popularStore *store.Store
}

func NewHandler(ls, ps *store.Store) *Handler {
	return &Handler{ls, ps}
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

// GetNews returns news grouped by source
func (h *Handler) GetNews(w http.ResponseWriter, r *http.Request) {
	limit := parseLimit(r.URL.Query().Get("limit"))
	news := h.latestStore.GetGrouped(limit)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status": "success",
		"items":  news,
	})
}

// GetPopular returns popular news grouped by source
func (h *Handler) GetPopular(w http.ResponseWriter, r *http.Request) {
	limit := parseLimit(r.URL.Query().Get("limit"))
	news := h.popularStore.GetGrouped(limit)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status": "success",
		"items":  news,
	})
}

func parseLimit(s string) int {
	if s == "" {
		return 0
	}

	limit, err := strconv.Atoi(s)
	if err != nil || limit < 1 {
		return 0
	}
	if limit > 20 {
		return 20
	}

	return limit
}
