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
	cfg := config.DefaultConfig()

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
	manager.Register(scraper.NewKatadataScraper(client))
	manager.Register(scraper.NewStockWatchScraper(client))
	manager.Register(scraper.NewIDXChannelScraper(client))
	manager.Register(scraper.NewKabarbursaScraper(client))
	manager.Register(scraper.NewKontanScraper(client))
	manager.Register(scraper.NewIDNFinansialsScraper(client))
	manager.Register(scraper.NewSindonewsScraper(client))
	manager.Register(scraper.NewCGSIScraper(client))
	manager.Register(scraper.NewLiputan6Scraper(client))
	manager.Register(scraper.NewInvestorIDScraper(client))

	go func() {
		scrapeNews(manager, newsStore)

		ticker := time.NewTicker(getPollingInterval())
		for range ticker.C {
			scrapeNews(manager, newsStore)
		}
	}()

	mux := http.NewServeMux()
	handler := api.NewHandler(newsStore)
	handler.RegisterRoutes(mux)

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

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)
}

func isMarketHours() bool {
	jakarta, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(jakarta)
	hour := now.Hour()

	return hour >= 9 && hour < 18
}

func getPollingInterval() time.Duration {
	if isMarketHours() {
		return 5 * time.Minute
	}
	return 60 * time.Minute
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
