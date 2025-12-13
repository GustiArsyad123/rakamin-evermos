package file

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Usecase interface {
	UploadFile(userID int64, file multipart.File, header *multipart.FileHeader) (string, error)
}

type fileUsecase struct {
	repo Repository
}

func NewUsecase(r Repository) Usecase {
	return &fileUsecase{repo: r}
}

func (u *fileUsecase) UploadFile(userID int64, file multipart.File, header *multipart.FileHeader) (string, error) {
	// Validate file type
	allowedTypes := map[string]bool{
		"image/jpeg":      true,
		"image/png":       true,
		"image/gif":       true,
		"image/webp":      true,
		"application/pdf": true,
		"text/plain":      true,
	}

	contentType := header.Header.Get("Content-Type")
	if !allowedTypes[contentType] {
		return "", fmt.Errorf("file type not allowed: %s", contentType)
	}

	// Validate file size (max 5MB)
	maxSize := int64(5 << 20) // 5MB
	if header.Size > maxSize {
		return "", fmt.Errorf("file too large: max 5MB")
	}

	// Create uploads directory
	uploadDir := "uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", err
	}

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		// Try to determine from content type
		switch contentType {
		case "image/jpeg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		case "image/gif":
			ext = ".gif"
		case "image/webp":
			ext = ".webp"
		case "application/pdf":
			ext = ".pdf"
		case "text/plain":
			ext = ".txt"
		}
	}

	// Sanitize filename
	safeName := strings.ReplaceAll(header.Filename, "..", "")
	safeName = strings.ReplaceAll(safeName, "/", "")
	safeName = strings.ReplaceAll(safeName, "\\", "")

	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("%d_%d_%s%s", userID, timestamp, safeName, ext)
	filePath := filepath.Join(uploadDir, filename)

	// Save file
	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		os.Remove(filePath) // cleanup on error
		return "", err
	}

	// Return relative URL
	return "/" + filePath, nil
}
