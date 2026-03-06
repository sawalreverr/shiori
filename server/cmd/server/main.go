package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
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

	slog.Info("Database initialized", "path", cfg.DatabasePath)

	newsStore := store.NewStoreWithDB(db)

	if err := newsStore.LoadFromDB(); err != nil {
			slog.Warn("Failed to load from database", "error", err)
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
	manager.Register(scraper.NewRepublikaScraper(client))
	manager.Register(scraper.NewAjaibScraper(client))

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
	Handler: api.CORSMiddleware(api.RequestIDMiddleware(mux)),
	}

	go func() {
			slog.Info("Server starting", "port", cfg.Port, "addr", fmt.Sprintf("http://localhost:%d", cfg.Port))
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
				slog.Error("Server failed to start", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server")
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
			slog.Error("Scraper error", "error", err)
	}

	count := 0
	for _, news := range newsArr {
		if newsStore.Save(news) {
			count++
		}
	}

	slog.Info("Scrape completed", "count", count, "market_hours", isMarketHours())
}
