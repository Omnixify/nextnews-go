package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramBotToken string
	GeminiApi        string
	RedisUrl         string
}

func Load() *Config {
	godotenv.Load()
	return &Config{
		TelegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		GeminiApi:        os.Getenv("GEMINI_API"),
		RedisUrl:         os.Getenv("REDIS_URL"),
	}
}
