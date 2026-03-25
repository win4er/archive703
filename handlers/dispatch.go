package handlers

import (
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func Dispatch(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
    if msg.IsCommand() {
        // Обработка команд
        switch msg.Command() {
        case "start":
            handleStart(bot, msg)
        case "help":
            handleHelp(bot, msg)
        case "upload":
            handleUpload(bot, msg)
        case "search":
            // извлечь аргументы
            handleSearch(bot, msg, args)
        }
        return
    }

    if msg.Document != nil {
        handleDocument(bot, msg)
        return
    }

    if msg.Photo != nil {
        handlePhoto(bot, msg)
        return
    }

    if msg.Text != "" {
        handleText(bot, msg)
    }
}
