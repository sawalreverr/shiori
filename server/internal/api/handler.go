package api

import (
	"encoding/json"
	"net/http"
	"shiori/internal/model"
	"shiori/internal/store"
	"strconv"
	"time"
)

type Handler struct {
	store *store.Store
}

func NewHandler(s *store.Store) *Handler {
	return &Handler{s}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.Health)
	mux.HandleFunc("/api/news", h.GetNews)
}

// Health returns server status
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func (h *Handler) GetNews(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "public, max-age=60")
	h.writeGrouped(w, r)
}

func (h *Handler) writeGrouped(w http.ResponseWriter, r *http.Request) {
	limit := parseLimit(r.URL.Query().Get("limit"))
	groups := h.store.GetGrouped(limit)
	lastScrapedAt := h.store.GetLastScrapedAt()

	respGroups := mapGroups(groups)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(model.MarketResponse{
		Status:        "success",
		LastScrapedAt: lastScrapedAt,
		SourceCount:   len(respGroups),
		TotalNews:     calculateTotalNews(respGroups),
		Items:         respGroups,
	})
}

func mapGroups(groups []store.SourceGroup) []model.SourceGroupResponse {
	out := make([]model.SourceGroupResponse, 0, len(groups))

	for _, g := range groups {
		respNews := make([]model.MarketNewsResponse, 0, len(g.News))
		for _, n := range g.News {
			respNews = append(respNews, model.MarketNewsResponse{
				Title:       n.Title,
				URL:         n.URL,
				ImageURL:    n.ImageURL,
				PublishedAt: n.PublishedAt,
			})
		}

		out = append(out, model.SourceGroupResponse{Source: g.Source, News: respNews})
	}

	return out
}

func calculateTotalNews(groups []model.SourceGroupResponse) int {
	total := 0
	for _, g := range groups {
		total += len(g.News)
	}
	return total
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
