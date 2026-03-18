package storage

import (
    "context"
    "fmt"
    "io"
    "archive703/models"
    
    "google.golang.org/api/drive/v3"
    "google.golang.org/api/option"
)

type GoogleDrive struct {
    service  *drive.Service
    folderID string
}

func NewGoogleDrive(credentialsFile, folderID string) (*GoogleDrive, error) {
    ctx := context.Background()
    
    service, err := drive.NewService(ctx, option.WithCredentialsFile(credentialsFile))
    if err != nil {
        return nil, fmt.Errorf("failed to create drive service: %v", err)
    }
    
    return &GoogleDrive{
        service:  service,
        folderID: folderID,
    }, nil
}

func (g *GoogleDrive) UploadFile(file io.Reader, filename string, mimeType string) (string, string, error) {
    fileMetadata := &drive.File{
        Name:    filename,
        Parents: []string{g.folderID},
    }
    
    uploadedFile, err := g.service.Files.Create(fileMetadata).Media(file).Do()
    if err != nil {
        return "", "", fmt.Errorf("failed to upload file: %v", err)
    }
    
    // Make file publicly accessible
    _, err = g.service.Permissions.Create(uploadedFile.Id, &drive.Permission{
        Type: "anyone",
        Role: "reader",
    }).Do()
    if err != nil {
        return uploadedFile.Id, "", fmt.Errorf("failed to set permissions: %v", err)
    }
    
    downloadLink := fmt.Sprintf("https://drive.google.com/uc?id=%s&export=download", uploadedFile.Id)
    
    return uploadedFile.Id, downloadLink, nil
}

func (g *GoogleDrive) DeleteFile(fileID string) error {
    err := g.service.Files.Delete(fileID).Do()
    if err != nil {
        return fmt.Errorf("failed to delete file: %v", err)
    }
    return nil
}

func (g *GoogleDrive) GetFileLink(fileID string) string {
    return fmt.Sprintf("https://drive.google.com/file/d/%s/view", fileID)
}

func (g *GoogleDrive) ListFiles() ([]*drive.File, error) {
    files, err := g.service.Files.List().
        Q(fmt.Sprintf("'%s' in parents", g.folderID)).
        Fields("files(id, name, mimeType, createdTime, size)").
        Do()
    if err != nil {
        return nil, fmt.Errorf("failed to list files: %v", err)
    }
    return files.Files, nil
}
