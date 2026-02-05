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

	newsStore := store.NewStoreWithDB(db)

	if err := newsStore.LoadFromDB(); err != nil {
		log.Printf("Warning: failed to load from DB: %v", err)
	}

	manager := scraper.NewManager(cfg)
	client := manager.GetHTTPClient()

	manager.Register(scraper.NewBloombergTechnozScraper(client))
	manager.Register(scraper.NewCNBCIndonesiaScraper(client))
	manager.Register(scraper.NewBisnisIndonesiaScraper(client))
	manager.Register(scraper.NewDetikFinanceScraper(client))

	go func() {
		scrapeNews(manager, newsStore)

		ticker := time.NewTicker(cfg.ScraperInterval)
		for range ticker.C {
			scrapeNews(manager, newsStore)
		}
	}()

	mux := http.NewServeMux()
	handler := api.NewHandler(newsStore)
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

func scrapeNews(manager *scraper.Manager, newsStore *store.Store) {
	ctx := context.Background()

	newsArr, errors := manager.ScrapeAll(ctx)
	for _, err := range errors {
		log.Printf("Error: %v", err)
	}

	count := 0
	for _, news := range newsArr {
		if newsStore.Save(news) {
			count++
		}
	}

	log.Printf("Scrape done: %d news items", count)
}
