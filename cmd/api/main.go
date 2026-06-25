package main

import (
	"context"
	"log"
	"net/http"

	"github.com/PooryaAlirezazadeh/TeleBot/internal/cache"
	"github.com/PooryaAlirezazadeh/TeleBot/internal/config"
	"github.com/PooryaAlirezazadeh/TeleBot/internal/scraper"
	"github.com/PooryaAlirezazadeh/TeleBot/internal/telegram"
	"github.com/PooryaAlirezazadeh/TeleBot/internal/translator"
)

func main() {
	cfg := config.Load()
	if cfg.OpenModelApi == "" || cfg.RedisUrl == "" || cfg.TelegramBotToken == "" {
		log.Fatal("Missing required environment variables: TELEGRAM_BOT_TOKEN or GEMINI_API_KEY or TELEGRAM_BOT_API")
	}
	scraper := scraper.New()
	translator := translator.New(cfg.OpenModelApi)

	cache, err := cache.New(cfg.RedisUrl)
	if err != nil {
		log.Fatalf("Failed to initialize cache package: %v", err)
	}
	engine := telegram.NewEngine(scraper, translator, cfg.TelegramBotToken, cache)
	engine.Start(context.Background())

	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Bot is running!"))
		})
		if err := http.ListenAndServe(":7860", nil); err != nil {
			log.Fatal(err)
		}
	}()
}
