package main

import (
    "log"
    "os"
    "os/signal"
    "syscall"

    "your-module/config"
    "your-module/handlers"
    "your-module/storage"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
    // 1. Загрузить конфиг (токен и т.д.)
    cfg := config.Load()

    // 2. Инициализировать БД (создать таблицы, если нет)
    db := storage.InitDB(cfg.DBPath)

    // 3. Создать HTTP-клиент (без прокси или с VPN, как сейчас)
    httpClient := &http.Client{...}

    // 4. Создать бота через NewBotAPIWithClient
    bot, err := tgbotapi.NewBotAPIWithClient(cfg.Token, tgbotapi.APIEndpoint, httpClient)
    if err != nil {
        log.Fatal("Failed to create bot:", err)
    }

    // 5. Настроить обновления (как у тебя сейчас)
    updateConfig := tgbotapi.NewUpdate(0)
    updateConfig.Timeout = 60
    updates := bot.GetUpdatesChan(updateConfig)

    // 6. Запустить обработчики (передать bot и db)
    handlers.Setup(bot, db)

    // 7. Обработка сигналов (Ctrl+C)
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

    // 8. Цикл обработки сообщений (можно вынести в handlers)
    go func() {
        for update := range updates {
            if update.Message == nil {
                continue
            }
            // диспетчеризация: команда, документ, фото, текст
            handlers.Dispatch(bot, update.Message)
        }
    }()

    <-stop
    log.Println("Bot stopped")
}
