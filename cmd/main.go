package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	// Определяем путь к корню проекта (где находится .env файл)
	// Из ./cmd/main.go нужно подняться на уровень выше
	projectRoot, err := os.Getwd()
	if err != nil {
		log.Fatal("Failed to get working directory:", err)
	}
	
	// Если мы в cmd, поднимаемся на уровень выше
	if filepath.Base(projectRoot) == "cmd" {
		projectRoot = filepath.Dir(projectRoot)
	}
	
	envPath := filepath.Join(projectRoot, ".env")
	
	// Загружаем переменные окружения
	err = godotenv.Load(envPath)
	if err != nil {
		log.Printf("Warning: .env file not found at %s, trying default location", envPath)
		// Пробуем загрузить из текущей директории
		err = godotenv.Load()
		if err != nil {
			log.Println("Warning: .env file not found, using environment variables")
		}
	}

	// Получаем токен бота
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("BOT_TOKEN is not set")
	}

	// Создаем HTTP клиент с увеличенными таймаутами
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSHandshakeTimeout:   15 * time.Second,
			ResponseHeaderTimeout: 15 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			IdleConnTimeout:       30 * time.Second,
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   10,
		},
	}

	// Создаем бота с кастомным HTTP клиентом
	bot, err := tgbotapi.NewBotAPIWithClient(token, tgbotapi.APIEndpoint, httpClient)
	if err != nil {
		log.Fatal("Failed to create bot:", err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Настройка обновлений
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60
	updates := bot.GetUpdatesChan(updateConfig)

	// Обработка сигналов
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Println("Bot started!")
		for update := range updates {
			if update.Message == nil {
				continue
			}

			chatID := update.Message.Chat.ID
			username := update.Message.From.UserName

			// Обработка команд
			if update.Message.IsCommand() {
				handleCommand(bot, update.Message)
				continue
			}

			// Обработка документов
			if update.Message.Document != nil {
				handleDocument(bot, update.Message)
				continue
			}

			// Обработка фото
			if update.Message.Photo != nil {
				handlePhoto(bot, update.Message)
				continue
			}

			// Обработка текста (эхо)
			if update.Message.Text != "" {
				log.Printf("Message from @%s (%d): %s", username, chatID, update.Message.Text)
				msg := tgbotapi.NewMessage(chatID, "Эхо: "+update.Message.Text)
				_, err := bot.Send(msg)
				if err != nil {
					log.Printf("Error sending message: %v", err)
				}
			}
		}
	}()

	<-stop
	log.Println("Bot stopped")
}

func handleCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	chatID := message.Chat.ID
	command := message.Command()

	switch command {
	case "start":
		msg := tgbotapi.NewMessage(chatID,
			"Привет! Я бот для архивации учебных материалов.\n\n"+
				"Доступные команды:\n"+
				"/start - показать это сообщение\n"+
				"/help - помощь\n"+
				"/upload - загрузить файл\n"+
				"/search - поиск по архиву")
		bot.Send(msg)

	case "help":
		msg := tgbotapi.NewMessage(chatID,
			"Отправьте мне файл (документ, фото, видео) и я сохраню его в архив.\n"+
				"Для поиска используйте /search <ключевое слово>")
		bot.Send(msg)

	case "upload":
		msg := tgbotapi.NewMessage(chatID, "Отправьте файл для загрузки в архив")
		bot.Send(msg)

	case "search":
		msg := tgbotapi.NewMessage(chatID, "Функция поиска в разработке")
		bot.Send(msg)
	}
}

func handleDocument(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	chatID := message.Chat.ID
	fileID := message.Document.FileID
	fileName := message.Document.FileName

	log.Printf("Received document: %s (ID: %s)", fileName, fileID)

	msg := tgbotapi.NewMessage(chatID,
		"Файл получен: "+fileName+"\n"+
			"ID: "+fileID+"\n"+
			"Размер: "+formatFileSize(int64(message.Document.FileSize)))
	bot.Send(msg)
}

func handlePhoto(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	chatID := message.Chat.ID

	photos := message.Photo
	if len(photos) > 0 {
		fileID := photos[len(photos)-1].FileID
		log.Printf("Received photo: %s", fileID)

		msg := tgbotapi.NewMessage(chatID, "Фото получено! ID: "+fileID)
		bot.Send(msg)
	}
}

func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
