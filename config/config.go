package config

type Config struct {
    Token  string
    DBPath string
    // другие настройки: AdminIDs, ArchivePath и т.д.
}

func Load() *Config {
    // Загрузить .env (можно через godotenv)
    // Прочитать BOT_TOKEN, DB_PATH, ADMIN_IDS
    // Вернуть заполненный Config
    return &Config{}
}
