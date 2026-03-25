package handlers

import (
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func handleText(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
    // Пока — эхо, как сейчас
    // В будущем — можно обрабатывать не-команды, например, предмет для загрузки
}
