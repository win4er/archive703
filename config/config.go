package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Token       string
	DBPath      string
	StoragePath string
	AdminIDs    []int64 // Telegram ID — лучше int64, так как ID большие
}

// Load загружает конфигурацию из .env и переменных окружения
func Load() *Config {
	cfg := &Config{}

	// Пробуем загрузить .env из корня проекта
	// (предполагаем, что config.go лежит в bot/config/, а .env на уровень выше)
	envPath := "../.env"
	err := godotenv.Load(envPath)
	if err != nil {
		log.Printf("Warning: .env file not found at %s, using system env", envPath)
	}

	cfg.Token = os.Getenv("BOT_TOKEN")
	if cfg.Token == "" {
		log.Fatal("BOT_TOKEN is not set")
	}

	cfg.DBPath = os.Getenv("DB_PATH")
	if cfg.DBPath == "" {
		log.Fatal("DB_PATH is not set")
	}

	cfg.StoragePath = os.Getenv("STORAGE_PATH")
	if cfg.StoragePath == "" {
		log.Fatal("STORAGE_PATH is not set")
	}

	adminsStr := os.Getenv("ADMINS_IDS")
	if adminsStr == "" {
		log.Println("ADMINS_IDS not set, no admin users")
	} else {
		parts := strings.Split(adminsStr, ",")
		cfg.AdminIDs = make([]int64, 0, len(parts))
		for _, part := range parts {
			id, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64)
			if err != nil {
				log.Printf("Warning: invalid admin ID %q: %v", part, err)
				continue
			}
			cfg.AdminIDs = append(cfg.AdminIDs, id)
		}
	}

	log.Printf("Config loaded: token=%s, db=%s, storage=%s, admins=%v",
		maskToken(cfg.Token), cfg.DBPath, cfg.StoragePath, cfg.AdminIDs)

	return cfg
}

// maskToken скрывает часть токена для логов (чтобы не светить его)
func maskToken(token string) string {
	if len(token) > 10 {
		return token[:5] + "..." + token[len(token)-3:]
	}
	return "***"
}
