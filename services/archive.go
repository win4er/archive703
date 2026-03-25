package services

import (
    "path/filepath"
)

// BuildFilePath формирует путь для сохранения файла
// Например: archive/703/теормех/username_2025-03-25_filename.pdf
func BuildFilePath(subject, filename, authorUsername string) string {
    // Создать папку, если нет
    // Вернуть полный путь
    return filepath.Join("archive", "703", subject, filename)
}
