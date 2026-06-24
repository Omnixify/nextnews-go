package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramBotToken string
	HuggingFaceApi   string
	RedisUrl         string
}

func Load() *Config {
	godotenv.Load()
	return &Config{
		TelegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		HuggingFaceApi:   os.Getenv("HUGGING_FACE_API"),
		RedisUrl:         os.Getenv("REDIS_URL"),
	}
}
