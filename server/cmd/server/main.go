package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"shiori/internal/config"
	"shiori/internal/scraper"
	"shiori/internal/store"
	"time"
)

func main() {
	// load config
	cfg := config.DefaultConfig()

	// create store
	newsStore := store.NewStore()

	// create scraper
	manager := scraper.NewManager(cfg)
	kompasScraper := scraper.NewKompasScraper(manager.GetHTTPClient())
	manager.Register(kompasScraper)

	go func() {
		scrapeNews(manager, newsStore)

		// then scrape every interval
		ticker := time.NewTicker(cfg.ScraperInterval)
		for range ticker.C {
			scrapeNews(manager, newsStore)
		}
	}()

	// setup HTTP routes
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	http.HandleFunc("/api/news", func(w http.ResponseWriter, r *http.Request) {
		news := newsStore.GetAll()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"items":  news,
			"count":  len(news),
		})
	})

	http.HandleFunc("/api/news/popular", func(w http.ResponseWriter, r *http.Request) {
		news := newsStore.GetPopular()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"items":  news,
			"count":  len(news),
		})
	})

	// start server
	server := &http.Server{
		Addr: fmt.Sprintf(":%d", cfg.Port),
	}

	go func() {
		log.Printf("Server running at http://localhost:%d", cfg.Port)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// ctrl+c
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)
}

func scrapeNews(manager *scraper.Manager, newsStore *store.Store) {
	ctx := context.Background()

	// Scrape latest
	newsArr, errors := manager.ScrapeAllLatest(ctx)
	for _, err := range errors {
		log.Printf("Error: %v", err)
	}

	// Save news
	newCount := 0
	for _, news := range newsArr {
		if newsStore.Save(news) {
			newCount++
		}
	}

	// Scrape Popular
	popular, errors := manager.ScrapeAllPopular(ctx)
	for _, err := range errors {
		log.Printf("Error: %v", err)
	}

	for _, news := range popular {
		if newsStore.Save(news) {
			newCount++
		}
	}

	log.Printf("Scrape done: %d news, %d total", newCount, newsStore.Count())
}
