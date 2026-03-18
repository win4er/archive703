package bot

import (
    "context"
    "fmt"
    "log"
    "strconv"
    "strings"
    "time"
    
    "archive703/models"
    "archive703/storage"
    
    "github.com/go-telegram/bot"
    "github.com/go-telegram/bot/models"
)

type ArchiveBot struct {
    bot         *bot.Bot
    db          *models.DB
    drive       *storage.GoogleDrive
    adminIDs    []int64
    userStates  map[int64]*UserState
}

type UserState struct {
    Step        string
    Material    *models.Material
    TempData    map[string]interface{}
}

func New(token string, db *models.DB, drive *storage.GoogleDrive, adminIDs []int64) (*ArchiveBot, error) {
    b, err := bot.New(token)
    if err != nil {
        return nil, err
    }
    
    archiveBot := &ArchiveBot{
        bot:        b,
        db:         db,
        drive:      drive,
        adminIDs:   adminIDs,
        userStates: make(map[int64]*UserState),
    }
    
    return archiveBot, nil
}

func (ab *ArchiveBot) Start() {
    ab.bot.RegisterHandler(bot.HandlerTypeMessageHandler, "", bot.MatchTypePrefix, ab.messageHandler)
    ab.bot.RegisterHandler(bot.HandlerTypeCallbackQueryData, "", bot.MatchTypePrefix, ab.callbackHandler)
    
    log.Println("Bot started")
    ab.bot.Start(context.Background())
}

func (ab *ArchiveBot) messageHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
    userID := update.Message.From.ID
    
    // Check/register user
    ab.ensureUser(update.Message.From)
    
    if update.Message.Text == "/start" {
        ab.handleStart(ctx, update)
        return
    }
    
    // Handle user state
    if state, exists := ab.userStates[userID]; exists {
        ab.handleState(ctx, update, state)
        return
    }
    
    // Default handler
    b.SendMessage(ctx, &bot.SendMessageParams{
        ChatID: update.Message.Chat.ID,
        Text:   "Используйте кнопки меню",
        ReplyMarkup: mainKeyboard(ab.isAdmin(userID)),
    })
}

func (ab *ArchiveBot) callbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
    callback := update.CallbackQuery
    userID := callback.From.ID
    
    ab.ensureUser(callback.From)
    
    data := callback.Data
    ab.bot.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
        CallbackQueryID: callback.ID,
    })
    
    switch {
    case data == "search":
        ab.handleSearch(ctx, callback)
    case data == "upload":
        ab.handleUploadStart(ctx, callback)
    case strings.HasPrefix(data, "type_"):
        materialType := strings.TrimPrefix(data, "type_")
        ab.handleMaterialType(ctx, callback, materialType)
    case data == "teachers":
        ab.handleTeachers(ctx, callback)
    case data == "my_rating":
        ab.handleMyRating(ctx, callback)
    case data == "admin":
        ab.handleAdmin(ctx, callback)
    case data == "admin_pending":
        ab.handleAdminPending(ctx, callback)
    case data == "admin_stats":
        ab.handleAdminStats(ctx, callback)
    case strings.HasPrefix(data, "approve_"):
        materialID := strings.TrimPrefix(data, "approve_")
        ab.handleApprove(ctx, callback, materialID)
    case strings.HasPrefix(data, "reject_"):
        materialID := strings.TrimPrefix(data, "reject_")
        ab.handleReject(ctx, callback, materialID)
    case data == "back_main":
        ab.handleBackMain(ctx, callback)
    case data == "cancel":
        delete(ab.userStates, userID)
        ab.sendMessage(ctx, callback.Message.Chat.ID, "Действие отменено")
    }
}

func (ab *ArchiveBot) handleStart(ctx context.Context, update *models.Update) {
    text := `📚 Добро пожаловать в Archive703!

Здесь вы можете найти и поделиться:
• Экзаменационными материалами
• Тестами
• Информацией о преподавателях
• Конспектами
• Сливами заданий

Выберите действие:`
    
    ab.bot.SendMessage(ctx, &bot.SendMessageParams{
        ChatID: update.Message.Chat.ID,
        Text:   text,
        ReplyMarkup: mainKeyboard(ab.isAdmin(update.Message.From.ID)),
    })
}

func (ab *ArchiveBot) handleUploadStart(ctx context.Context, callback *models.CallbackQuery) {
    userID := callback.From.ID
    
    ab.userStates[userID] = &UserState{
        Step: "waiting_type",
        Material: &models.Material{
            UserID:   userID,
            UserName: callback.From.Username,
        },
    }
    
    ab.bot.SendMessage(ctx, &bot.SendMessageParams{
        ChatID: callback.Message.Chat.ID,
        Text:   "Выберите тип материала:",
        ReplyMarkup: materialTypeKeyboard(),
    })
}

func (ab *ArchiveBot) handleMaterialType(ctx context.Context, callback *models.CallbackQuery, materialType string) {
    userID := callback.From.ID
    
    if state, exists := ab.userStates[userID]; exists {
        state.Material.Type = models.MaterialType(materialType)
        state.Step = "waiting_title"
        
        ab.bot.SendMessage(ctx, &bot.SendMessageParams{
            ChatID: callback.Message.Chat.ID,
            Text:   "Введите название материала:",
        })
    }
}

func (ab *ArchiveBot) handleState(ctx context.Context, update *models.Update, state *UserState) {
    userID := update.Message.From.ID
    
    switch state.Step {
    case "waiting_title":
        state.Material.Title = update.Message.Text
        state.Step = "waiting_description"
        
        ab.bot.SendMessage(ctx, &bot.SendMessageParams{
            ChatID: update.Message.Chat.ID,
            Text:   "Введите описание материала:",
        })
        
    case "waiting_description":
        state.Material.Description = update.Message.Text
        state.Step = "waiting_file"
        
        ab.bot.SendMessage(ctx, &bot.SendMessageParams{
            ChatID: update.Message.Chat.ID,
            Text:   "Отправьте файл (PDF, изображение или документ):",
        })
        
    case "waiting_file":
        if update.Message.Document == nil && update.Message.Photo == nil {
            ab.bot.SendMessage(ctx, &bot.SendMessageParams{
                ChatID: update.Message.Chat.ID,
                Text:   "Пожалуйста, отправьте файл",
            })
            return
        }
        
        ab.processUpload(ctx, update, state)
    }
}

func (ab *ArchiveBot) processUpload(ctx context.Context, update *models.Update, state *UserState) {
    userID := update.Message.From.ID
    var fileID, fileName string
    var fileSize int
    
    if update.Message.Document != nil {
        fileID = update.Message.Document.FileID
        fileName = update.Message.Document.FileName
        fileSize = update.Message.Document.FileSize
    } else if update.Message.Photo != nil {
        photo := update.Message.Photo[len(update.Message.Photo)-1]
        fileID = photo.FileID
        fileName = fmt.Sprintf("photo_%d.jpg", time.Now().Unix())
        fileSize = photo.FileSize
    }
    
    // Download file from Telegram
    fileData, err := ab.downloadTelegramFile(ctx, fileID)
    if err != nil {
        ab.bot.SendMessage(ctx, &bot.SendMessageParams{
            ChatID: update.Message.Chat.ID,
            Text:   "Ошибка при загрузке файла. Попробуйте еще раз.",
        })
        delete(ab.userStates, userID)
        return
    }
    
    // Upload to Google Drive
    driveID, driveLink, err := ab.drive.UploadFile(fileData, fileName, "application/octet-stream")
    if err != nil {
        ab.bot.SendMessage(ctx, &bot.SendMessageParams{
            ChatID: update.Message.Chat.ID,
            Text:   "Ошибка при сохранении файла. Попробуйте позже.",
        })
        delete(ab.userStates, userID)
        return
    }
    
    state.Material.FileID = fileID
    state.Material.DriveFileID = driveID
    state.Material.DriveLink = driveLink
    state.Material.Status = "pending"
    state.Material.CreatedAt = time.Now()
    
    // Save to database
    err = ab.saveMaterial(state.Material)
    if err != nil {
        ab.bot.SendMessage(ctx, &bot.SendMessageParams{
            ChatID: update.Message.Chat.ID,
            Text:   "Ошибка при сохранении в базу данных.",
        })
        delete(ab.userStates, userID)
        return
    }
    
    // Notify user
    ab.bot.SendMessage(ctx, &bot.SendMessageParams{
        ChatID: update.Message.Chat.ID,
        Text:   "✅ Материал отправлен на проверку администратору!",
    })
    
    // Notify admins
    ab.notifyAdmins(ctx, state.Material)
    
    delete(ab.userStates, userID)
}

func (ab *ArchiveBot) handleAdminPending(ctx context.Context, callback *models.CallbackQuery) {
    if !ab.isAdmin(callback.From.ID) {
        return
    }
    
    materials, err := ab.getPendingMaterials()
    if err != nil {
        ab.bot.SendMessage(ctx, &bot.SendMessageParams{
            ChatID: callback.Message.Chat.ID,
            Text:   "Ошибка при загрузке материалов",
        })
        return
    }
    
    if len(materials) == 0 {
        ab.bot.SendMessage(ctx, &bot.SendMessageParams{
            ChatID: callback.Message.Chat.ID,
            Text:   "Нет материалов на проверке",
        })
        return
    }
    
    for _, m := range materials {
        text := fmt.Sprintf(
            "📥 Новый материал на проверку\n\n"+
            "Тип: %s\n"+
            "Название: %s\n"+
            "Описание: %s\n"+
            "От: @%s\n"+
            "Дата: %s",
            m.Type, m.Title, m.Description, m.UserName, m.CreatedAt.Format("02.01.2006 15:04"),
        )
        
        ab.bot.SendMessage(ctx, &bot.SendMessageParams{
            ChatID: callback.Message.Chat.ID,
            Text:   text,
            ReplyMarkup: moderationKeyboard(m.ID),
        })
    }
}

func (ab *ArchiveBot) handleApprove(ctx context.Context, callback *models.CallbackQuery, materialID string) {
    if !ab.isAdmin(callback.From.ID) {
        return
    }
    
    err := ab.approveMaterial(materialID, callback.From.ID)
    if err != nil {
        ab.bot.SendMessage(ctx, &bot.SendMessageParams{
            ChatID: callback.Message.Chat.ID,
            Text:   "Ошибка при подтверждении",
        })
        return
    }
    
    ab.bot.SendMessage(ctx, &bot.SendMessageParams{
        ChatID: callback.Message.Chat.ID,
        Text:   "✅ Материал подтвержден и доступен для всех!",
    })
    
    // Update message
    ab.bot.EditMessageText(ctx, &bot.EditMessageTextParams{
        ChatID:    callback.Message.Chat.ID,
        MessageID: callback.Message.ID,
        Text:      callback.Message.Text + "\n\n✅ Подтверждено",
    })
}

func (ab *ArchiveBot) handleAdminStats(ctx context.Context, callback *models.CallbackQuery) {
    if !ab.isAdmin(callback.From.ID) {
        return
    }
    
    stats, err := ab.getStats()
    if err != nil {
        ab.bot.SendMessage(ctx, &bot.SendMessageParams{
            ChatID: callback.Message.Chat.ID,
            Text:   "Ошибка при загрузке статистики",
        })
        return
    }
    
    text := fmt.Sprintf(
        "📊 Статистика бота\n\n"+
        "👥 Всего пользователей: %d\n"+
        "📚 Всего материалов: %d\n"+
        "⏳ Ожидают проверки: %d\n"+
        "✅ Подтверждено: %d\n"+
        "👨‍🏫 Преподавателей в базе: %d",
        stats.TotalUsers, stats.TotalMaterials, stats.PendingMaterials,
        stats.ApprovedMaterials, stats.TotalTeachers,
    )
    
    ab.bot.SendMessage(ctx, &bot.SendMessageParams{
        ChatID: callback.Message.Chat.ID,
        Text:   text,
    })
}

// Helper functions
func (ab *ArchiveBot) isAdmin(userID int64) bool {
    for _, id := range ab.adminIDs {
        if id == userID {
            return true
        }
    }
    return false
}

func (ab *ArchiveBot) ensureUser(user *models.User) {
    // Check if user exists in DB, if not - create
    var exists bool
    err := ab.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = ?)", user.ID).Scan(&exists)
    if err != nil {
        log.Printf("Error checking user: %v", err)
        return
    }
    
    if !exists {
        _, err = ab.db.Exec(
            "INSERT INTO users (id, username, first_name, last_name, joined_at) VALUES (?, ?, ?, ?, ?)",
            user.ID, user.Username, user.FirstName, user.LastName, time.Now(),
        )
        if err != nil {
            log.Printf("Error creating user: %v", err)
        }
    }
}

func (ab *ArchiveBot) downloadTelegramFile(ctx context.Context, fileID string) ([]byte, error) {
    // Implement file download from Telegram
    // This requires using bot.FileDownload
    return []byte{}, nil
}

func (ab *ArchiveBot) saveMaterial(material *models.Material) error {
    material.ID = fmt.Sprintf("mat_%d", time.Now().UnixNano())
    
    _, err := ab.db.Exec(
        `INSERT INTO materials (id, user_id, user_name, type, title, description, 
                               file_id, drive_file_id, drive_link, status, created_at)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
        material.ID, material.UserID, material.UserName, material.Type, material.Title,
        material.Description, material.FileID, material.DriveFileID, material.DriveLink,
        material.Status, material.CreatedAt,
    )
    return err
}

func (ab *ArchiveBot) getPendingMaterials() ([]models.Material, error) {
    rows, err := ab.db.Query(
        "SELECT id, user_id, user_name, type, title, description, created_at FROM materials WHERE status = 'pending'",
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var materials []models.Material
    for rows.Next() {
        var m models.Material
        err := rows.Scan(&m.ID, &m.UserID, &m.UserName, &m.Type, &m.Title, &m.Description, &m.CreatedAt)
        if err != nil {
            continue
        }
        materials = append(materials, m)
    }
    return materials, nil
}

func (ab *ArchiveBot) approveMaterial(id string, adminID int64) error {
    now := time.Now()
    _, err := ab.db.Exec(
        "UPDATE materials SET status = 'approved', approved_at = ?, approved_by = ? WHERE id = ?",
        now, adminID, id,
    )
    return err
}

type Stats struct {
    TotalUsers       int
    TotalMaterials   int
    PendingMaterials int
    ApprovedMaterials int
    TotalTeachers    int
}

func (ab *ArchiveBot) getStats() (*Stats, error) {
    stats := &Stats{}
    
    err := ab.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&stats.TotalUsers)
    if err != nil {
        return nil, err
    }
    
    err = ab.db.QueryRow("SELECT COUNT(*) FROM materials").Scan(&stats.TotalMaterials)
    if err != nil {
        return nil, err
    }
    
    err = ab.db.QueryRow("SELECT COUNT(*) FROM materials WHERE status = 'pending'").Scan(&stats.PendingMaterials)
    if err != nil {
        return nil, err
    }
    
    err = ab.db.QueryRow("SELECT COUNT(*) FROM materials WHERE status = 'approved'").Scan(&stats.ApprovedMaterials)
    if err != nil {
        return nil, err
    }
    
    err = ab.db.QueryRow("SELECT COUNT(*) FROM teachers").Scan(&stats.TotalTeachers)
    if err != nil {
        return nil, err
    }
    
    return stats, nil
}

func (ab *ArchiveBot) notifyAdmins(ctx context.Context, material *models.Material) {
    text := fmt.Sprintf(
        "🆕 Новый материал на проверку!\n\n"+
        "От: @%s\n"+
        "Тип: %s\n"+
        "Название: %s\n\n"+
        "Зайдите в админ панель для проверки.",
        material.UserName, material.Type, material.Title,
    )
    
    for _, adminID := range ab.adminIDs {
        ab.bot.SendMessage(ctx, &bot.SendMessageParams{
            ChatID: adminID,
            Text:   text,
        })
    }
}

// Additional handlers to implement
func (ab *ArchiveBot) handleSearch(ctx context.Context, callback *models.CallbackQuery) {
    // Implement search functionality
}

func (ab *ArchiveBot) handleTeachers(ctx context.Context, callback *models.CallbackQuery) {
    // Implement teachers list
}

func (ab *ArchiveBot) handleMyRating(ctx context.Context, callback *models.CallbackQuery) {
    // Implement user rating
}

func (ab *ArchiveBot) handleAdmin(ctx context.Context, callback *models.CallbackQuery) {
    ab.bot.SendMessage(ctx, &bot.SendMessageParams{
        ChatID: callback.Message.Chat.ID,
        Text:   "⚙️ Панель администратора",
        ReplyMarkup: adminKeyboard(),
    })
}

func (ab *ArchiveBot) handleReject(ctx context.Context, callback *models.CallbackQuery, materialID string) {
    // Implement reject logic
}

func (ab *ArchiveBot) handleBackMain(ctx context.Context, callback *models.CallbackQuery) {
    ab.bot.SendMessage(ctx, &bot.SendMessageParams{
        ChatID: callback.Message.Chat.ID,
        Text:   "Главное меню",
        ReplyMarkup: mainKeyboard(ab.isAdmin(callback.From.ID)),
    })
}

func (ab *ArchiveBot) sendMessage(ctx context.Context, chatID int64, text string) {
    ab.bot.SendMessage(ctx, &bot.SendMessageParams{
        ChatID: chatID,
        Text:   text,
    })
}
