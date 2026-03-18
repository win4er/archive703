package main

import (
    "log"
    "archive703/bot"
    "archive703/config"
    "archive703/models"
    "archive703/storage"
)

func main() {
    // Load configuration
    cfg := config.Load()
    
    // Initialize database
    db, err := models.InitDB(cfg.DatabasePath)
    if err != nil {
        log.Fatal("Failed to initialize database:", err)
    }
    defer db.Close()
    
    // Initialize Google Drive storage
    drive, err := storage.NewGoogleDrive(cfg.GoogleCreds, cfg.DriveFolderID)
    if err != nil {
        log.Fatal("Failed to initialize Google Drive:", err)
    }
    
    // Create and start bot
    archiveBot, err := bot.New(cfg.BotToken, db, drive, cfg.AdminIDs)
    if err != nil {
        log.Fatal("Failed to create bot:", err)
    }
    
    archiveBot.Start()
}
