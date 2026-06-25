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
	// --- ADD THIS LOGGING BLOCK ---
	log.Println("=========================================")
	log.Println("Checking Environment Variables...")
	log.Printf("TELEGRAM_BOT_TOKEN status: Length = %d", len(cfg.TelegramBotToken))
	if len(cfg.TelegramBotToken) > 10 {
		log.Printf("TELEGRAM_BOT_TOKEN partial mask: %s...%s", cfg.TelegramBotToken[:5], cfg.TelegramBotToken[len(cfg.TelegramBotToken)-5:])
	} else {
		log.Println("WARNING: TELEGRAM_BOT_TOKEN is empty or too short!")
	}

	log.Printf("OPENMODEL_API status: Length = %d", len(cfg.OpenModelApi))
	log.Printf("REDIS_URL status: Length = %d", len(cfg.RedisUrl))
	log.Println("=========================================")
	if cfg.OpenModelApi == "" || cfg.RedisUrl == "" || cfg.TelegramBotToken == "" {
		log.Fatal("Missing required environment variables: TELEGRAM_BOT_TOKEN or GEMINI_API_KEY or TELEGRAM_BOT_API")
	}
	scraper := scraper.New()
	translator := translator.New(cfg.OpenModelApi)

	cache, err := cache.New(cfg.RedisUrl)
	if err != nil {
		log.Fatalf("Failed to initialize cache package: %v", err)
	}

	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Bot is running!"))
		})
		if err := http.ListenAndServe(":7860", nil); err != nil {
			log.Fatal(err)
		}
	}()

	engine := telegram.NewEngine(scraper, translator, cfg.TelegramBotToken, cache)
	engine.Start(context.Background())

}
