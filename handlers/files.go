package handlers

import (
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func handleDocument(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
    // 1. Получить fileID, fileName, fileSize
    // 2. Скачать файл (вызвать services.DownloadFile)
    // 3. Сохранить метаданные в БД (storage.SaveMaterial)
    // 4. Ответить пользователю
}

func handlePhoto(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
    // Аналогично, но для фото (можно брать самое большое)
}
