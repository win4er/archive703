package config

import (
    "log"
    "os"
    "github.com/joho/godotenv"
)

type Config struct {
    BotToken     string
    AdminIDs     []int64
    GoogleCreds  string
    DriveFolderID string
    DatabasePath string
}

func Load() *Config {
    err := godotenv.Load()
    if err != nil {
        log.Println("No .env file found, using environment variables")
    }

    return &Config{
        BotToken:     getEnv("BOT_TOKEN", ""),
        AdminIDs:     parseAdminIDs(getEnv("ADMIN_IDS", "")),
        GoogleCreds:  getEnv("GOOGLE_CREDS", "credentials.json"),
        DriveFolderID: getEnv("DRIVE_FOLDER_ID", ""),
        DatabasePath: getEnv("DB_PATH", "archive703.db"),
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func parseAdminIDs(ids string) []int64 {
    // Implementation for parsing admin IDs
    return []int64{}
}
