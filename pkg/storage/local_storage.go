package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type localStorage struct {
    basePath string
    baseURL  string
}

// NewLocalStorage creates new local storage instance
func NewLocalStorage(basePath, baseURL string) Storage {
    return &localStorage{
        basePath: basePath,
        baseURL:  baseURL,
    }
}

func (s *localStorage) Save(ctx context.Context, file multipart.File, header *multipart.FileHeader, dir string) (*FileInfo, error) {
    // Create directory if not exists
    uploadDir := filepath.Join(s.basePath, dir)
    if err := os.MkdirAll(uploadDir, 0755); err != nil {
        return nil, fmt.Errorf("failed to create directory: %w", err)
    }

    // Generate unique filename
    ext := filepath.Ext(header.Filename)
    filename := fmt.Sprintf("%s_%d%s", uuid.New().String(), time.Now().Unix(), ext)
    
    // Full path
    fullPath := filepath.Join(uploadDir, filename)
    
    // Create destination file
    dst, err := os.Create(fullPath)
    if err != nil {
        return nil, fmt.Errorf("failed to create file: %w", err)
    }
    defer dst.Close()

    // Copy file content
    size, err := io.Copy(dst, file)
    if err != nil {
        return nil, fmt.Errorf("failed to save file: %w", err)
    }

    // Relative path for database
    relativePath := filepath.Join(dir, filename)
    
    // Public URL
    url := s.GetURL(relativePath)

    return &FileInfo{
        Filename:     filename,
        Path:         relativePath,
        URL:          url,
        Size:         size,
        MimeType:     header.Header.Get("Content-Type"),
        OriginalName: header.Filename,
    }, nil
}

func (s *localStorage) Delete(ctx context.Context, path string) error {
    fullPath := filepath.Join(s.basePath, path)
    
    if err := os.Remove(fullPath); err != nil {
        if os.IsNotExist(err) {
            return nil // File already deleted, no error
        }
        return fmt.Errorf("failed to delete file: %w", err)
    }
    
    return nil
}

func (s *localStorage) GetURL(path string) string {
    // Normalize path separators for URL
    urlPath := filepath.ToSlash(path)
    return fmt.Sprintf("%s/%s", strings.TrimSuffix(s.baseURL, "/"), urlPath)
}

func (s *localStorage) Exists(ctx context.Context, path string) bool {
    fullPath := filepath.Join(s.basePath, path)
    _, err := os.Stat(fullPath)
    return err == nil
}