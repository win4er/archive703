package handlers

import (
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func handleStart(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
    // Ответить на /start
}

func handleHelp(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
    // Ответить на /help
}

func handleUpload(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
    // Ответить "Отправьте файл"
}

func handleSearch(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, args string) {
    // Вызвать storage.SearchMaterials, вывести результат
}
