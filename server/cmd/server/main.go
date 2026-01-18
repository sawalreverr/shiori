package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"shiori/internal/api"
	"shiori/internal/config"
	"shiori/internal/scraper"
	"shiori/internal/store"
)

func main() {
	// load config
	cfg := config.DefaultConfig()

	// Initialize SQLite database
	db, err := store.OpenDB(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	log.Printf("Database initialized at %s", cfg.DatabasePath)

	// create stores with DB backing
	latestStore := store.NewStoreWithDB(db, store.FeedTypeLatest)
	popularStore := store.NewStoreWithDB(db, store.FeedTypePopular)

	// Load existing data from DB into cache
	if err := latestStore.LoadFromDB(); err != nil {
		log.Printf("Warning: failed to load latest from DB: %v", err)
	}
	if err := popularStore.LoadFromDB(); err != nil {
		log.Printf("Warning: failed to load popular from DB: %v", err)
	}

	// create scraper
	manager := scraper.NewManager(cfg)
	client := manager.GetHTTPClient()

	manager.Register(scraper.NewKompasScraper(client))
	manager.Register(scraper.NewDetikScraper(client))
	manager.Register(scraper.NewBloomberTechnozScraper(client))
	manager.Register(scraper.NewLiputan6Scraper(client))
	manager.Register(scraper.NewTribunNewsScraper(client))
	manager.Register(scraper.NewCNNIndonesiaScraper(client))

	go func() {
		scrapeNews(manager, latestStore, popularStore)

		// then scrape every interval
		ticker := time.NewTicker(cfg.ScraperInterval)
		for range ticker.C {
			scrapeNews(manager, latestStore, popularStore)
		}
	}()

	// setup HTTP routes
	mux := http.NewServeMux()
	handler := api.NewHandler(latestStore, popularStore)
	handler.RegisterRoutes(mux)

	// start server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: api.CORSMiddleware(mux),
	}

	go func() {
		log.Printf("Server running at http://localhost:%d", cfg.Port)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// ctrl+c
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)
}

func scrapeNews(manager *scraper.Manager, latestStore, popularStore *store.Store) {
	ctx := context.Background()

	// Scrape latest
	newsArr, errors := manager.ScrapeAllLatest(ctx)
	for _, err := range errors {
		log.Printf("Error: %v", err)
	}

	// Save news
	latestCount := 0
	for _, news := range newsArr {
		if latestStore.Save(news) {
			latestCount++
		}
	}

	// Scrape Popular
	popular, errors := manager.ScrapeAllPopular(ctx)
	for _, err := range errors {
		log.Printf("Error: %v", err)
	}

	popularCount := 0
	for _, news := range popular {
		if popularStore.Save(news) {
			popularCount++
		}
	}

	log.Printf("Scrape done: %d latest, %d popular", latestCount, popularCount)
}
