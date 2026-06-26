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

	if cfg.OpenModelApi == "" || cfg.RedisUrl == "" || cfg.TelegramBotToken == "" {
		log.Fatal("Missing required environment variables")
	}

	// Initialize dependencies
	scrp := scraper.New()
	trans := translator.New(cfg.OpenModelApi)
	cch, err := cache.New(cfg.RedisUrl)
	if err != nil {
		log.Fatalf("Cache error: %v", err)
	}

	// Run the scraper once and exit
	engine := telegram.NewEngine(scrp, trans, cfg.TelegramBotToken, cch)
	engine.RunOnce(context.Background())
}
