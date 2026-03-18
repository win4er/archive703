package bot

import (
    "fmt"
    "archive703/models"
    
    "github.com/go-telegram/bot/models"
)

func mainKeyboard(isAdmin bool) *models.InlineKeyboardMarkup {
    buttons := [][]models.InlineKeyboardButton{
        {
            {Text: "📚 Найти материалы", CallbackData: "search"},
            {Text: "📤 Загрузить", CallbackData: "upload"},
        },
        {
            {Text: "👨‍🏫 Преподаватели", CallbackData: "teachers"},
            {Text: "⭐ Мой рейтинг", CallbackData: "my_rating"},
        },
    }
    
    if isAdmin {
        buttons = append(buttons, []models.InlineKeyboardButton{
            {Text: "⚙️ Админ панель", CallbackData: "admin"},
        })
    }
    
    return &models.InlineKeyboardMarkup{InlineKeyboard: buttons}
}

func materialTypeKeyboard() *models.InlineKeyboardMarkup {
    return &models.InlineKeyboardMarkup{
        InlineKeyboard: [][]models.InlineKeyboardButton{
            {
                {Text: "📝 Экзамены", CallbackData: "type_exam"},
                {Text: "✍️ Тесты", CallbackData: "type_test"},
            },
            {
                {Text: "👨‍🏫 Преподаватели", CallbackData: "type_teacher"},
                {Text: "📓 Конспекты", CallbackData: "type_notes"},
            },
            {
                {Text: "💥 Сливы", CallbackData: "type_leak"},
                {Text: "❌ Отмена", CallbackData: "cancel"},
            },
        },
    }
}

func adminKeyboard() *models.InlineKeyboardMarkup {
    return &models.InlineKeyboardMarkup{
        InlineKeyboard: [][]models.InlineKeyboardButton{
            {
                {Text: "⏳ Ожидают проверки", CallbackData: "admin_pending"},
                {Text: "✅ Подтвержденные", CallbackData: "admin_approved"},
            },
            {
                {Text: "📊 Статистика", CallbackData: "admin_stats"},
                {Text: "👥 Пользователи", CallbackData: "admin_users"},
            },
            {
                {Text: "🔙 Назад", CallbackData: "back_main"},
            },
        },
    }
}

func moderationKeyboard(materialID string) *models.InlineKeyboardMarkup {
    return &models.InlineKeyboardMarkup{
        InlineKeyboard: [][]models.InlineKeyboardButton{
            {
                {Text: "✅ Подтвердить", CallbackData: fmt.Sprintf("approve_%s", materialID)},
                {Text: "❌ Отклонить", CallbackData: fmt.Sprintf("reject_%s", materialID)},
            },
        },
    }
}
