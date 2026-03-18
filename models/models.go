package models

import (
    "time"
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
)

type MaterialType string

const (
    TypeExam       MaterialType = "exam"
    TypeTest       MaterialType = "test"
    TypeTeacher    MaterialType = "teacher"
    TypeNotes      MaterialType = "notes"
    TypeLeak       MaterialType = "leak"
)

type Material struct {
    ID          string       `json:"id"`
    UserID      int64        `json:"user_id"`
    UserName    string       `json:"user_name"`
    Type        MaterialType `json:"type"`
    Title       string       `json:"title"`
    Description string       `json:"description"`
    FileID      string       `json:"file_id"`      // Telegram file ID
    DriveFileID string       `json:"drive_file_id"` // Google Drive file ID
    DriveLink   string       `json:"drive_link"`
    Status      string       `json:"status"`       // pending, approved, rejected
    CreatedAt   time.Time    `json:"created_at"`
    ApprovedAt  *time.Time   `json:"approved_at"`
    ApprovedBy  int64        `json:"approved_by"`
    Views       int          `json:"views"`
    Likes       int          `json:"likes"`
    Dislikes    int          `json:"dislikes"`
}

type Teacher struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Department  string    `json:"department"`
    Subject     string    `json:"subject"`
    Rating      float64   `json:"rating"`
    Description string    `json:"description"`
}

type User struct {
    ID           int64     `json:"id"`
    Username     string    `json:"username"`
    FirstName    string    `json:"first_name"`
    LastName     string    `json:"last_name"`
    IsAdmin      bool      `json:"is_admin"`
    Rating       int       `json:"rating"`
    MaterialsCnt int       `json:"materials_cnt"`
    JoinedAt     time.Time `json:"joined_at"`
}

type DB struct {
    *sql.DB
}

func InitDB(dbPath string) (*DB, error) {
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return nil, err
    }

    // Create tables
    queries := []string{
        `CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY,
            username TEXT,
            first_name TEXT,
            last_name TEXT,
            is_admin BOOLEAN DEFAULT FALSE,
            rating INTEGER DEFAULT 0,
            materials_cnt INTEGER DEFAULT 0,
            joined_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )`,
        `CREATE TABLE IF NOT EXISTS materials (
            id TEXT PRIMARY KEY,
            user_id INTEGER,
            user_name TEXT,
            type TEXT,
            title TEXT,
            description TEXT,
            file_id TEXT,
            drive_file_id TEXT,
            drive_link TEXT,
            status TEXT DEFAULT 'pending',
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            approved_at DATETIME,
            approved_by INTEGER,
            views INTEGER DEFAULT 0,
            likes INTEGER DEFAULT 0,
            dislikes INTEGER DEFAULT 0,
            FOREIGN KEY(user_id) REFERENCES users(id)
        )`,
        `CREATE TABLE IF NOT EXISTS teachers (
            id TEXT PRIMARY KEY,
            name TEXT,
            department TEXT,
            subject TEXT,
            rating REAL DEFAULT 0,
            description TEXT
        )`,
        `CREATE TABLE IF NOT EXISTS feedback (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            user_id INTEGER,
            message TEXT,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY(user_id) REFERENCES users(id)
        )`,
    }

    for _, query := range queries {
        _, err = db.Exec(query)
        if err != nil {
            return nil, err
        }
    }

    return &DB{db}, nil
}
