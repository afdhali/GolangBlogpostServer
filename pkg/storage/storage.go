package storage

import (
	"context"
	"mime/multipart"
)

// FileInfo represents uploaded file information
type FileInfo struct {
    Filename     string
    Path         string
    URL          string
    Size         int64
    MimeType     string
    OriginalName string
}

// Storage interface for file storage operations
type Storage interface {
    // Save file and return file info
    Save(ctx context.Context, file multipart.File, header *multipart.FileHeader, dir string) (*FileInfo, error)
    
    // Delete file by path
    Delete(ctx context.Context, path string) error
    
    // Get file URL
    GetURL(path string) string
    
    // Check if file exists
    Exists(ctx context.Context, path string) bool
}