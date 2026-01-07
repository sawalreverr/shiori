package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"shiori/internal/api"
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
	client := manager.GetHTTPClient()

	manager.Register(scraper.NewKompasScraper(client))
	manager.Register(scraper.NewDetikScraper(client))

	go func() {
		scrapeNews(manager, newsStore)

		// then scrape every interval
		ticker := time.NewTicker(cfg.ScraperInterval)
		for range ticker.C {
			scrapeNews(manager, newsStore)
		}
	}()

	// setup HTTP routes
	mux := http.NewServeMux()
	handler := api.NewHandler(newsStore)
	handler.RegisterRoutes(mux)

	// start server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: mux,
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
