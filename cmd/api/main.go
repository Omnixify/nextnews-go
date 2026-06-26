package main

import (
	"context"
	"log"

	"github.com/PooryaAlirezazadeh/TeleBot/internal/cache"
	"github.com/PooryaAlirezazadeh/TeleBot/internal/config"
	"github.com/PooryaAlirezazadeh/TeleBot/internal/scraper"
	"github.com/PooryaAlirezazadeh/TeleBot/internal/telegram"
	"github.com/PooryaAlirezazadeh/TeleBot/internal/translator"
)

func main() {
	cfg := config.Load()

	if cfg.GeminiApi == "" || cfg.RedisUrl == "" || cfg.TelegramBotToken == "" {
		log.Fatal("Missing required environment variables")
	}

	// Initialize dependencies
	scrp := scraper.New()
	trans, err := translator.New(cfg.GeminiApi)
	if err != nil {
		log.Fatalf("Gemini api error : %v", err)
	}
	cch, err := cache.New(cfg.RedisUrl)
	if err != nil {
		log.Fatalf("Cache error: %v", err)
	}

	// Run the scraper once and exit
	engine := telegram.NewEngine(scrp, trans, cfg.TelegramBotToken, cch)
	engine.RunOnce(context.Background())
}
