package utils

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"siro-backend/constants" // Import constants
	"strings"
	"time"
)

// UploadConfig struct untuk konfigurasi upload
type UploadConfig struct {
	AllowedExts []string
	MaxFileSize int64
	SubDir      string
}

func DefaultImageConfig(subDir string) UploadConfig {
	return UploadConfig{
		AllowedExts: []string{".jpg", ".jpeg", ".png"},
		MaxFileSize: 2 * 1024 * 1024, // 2MB
		SubDir:      subDir,
	}
}

func SaveUploadedFile(file *multipart.FileHeader, config UploadConfig) (string, error) {
	// Validasi Ukuran
	if file.Size > config.MaxFileSize {
		return "", fmt.Errorf("file too large (max 2MB)")
	}

	// Validasi Ekstensi
	ext := strings.ToLower(filepath.Ext(file.Filename))
	valid := false
	for _, e := range config.AllowedExts {
		if e == ext {
			valid = true
			break
		}
	}
	if !valid {
		return "", fmt.Errorf("invalid file extension")
	}

	// Buat Nama File Unik
	newFileName := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), config.SubDir, ext)

	// --- GUNAKAN CONSTANT UNTUK ROOT PATH ---
	uploadPath := filepath.Join(constants.DirUploads, config.SubDir)

	// Ensure Directory Exists
	if err := os.MkdirAll(uploadPath, os.ModePerm); err != nil {
		return "", err
	}

	dst := filepath.Join(uploadPath, newFileName)

	// Save File
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	out, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err = io.Copy(out, src); err != nil {
		return "", err
	}

	// Return relative URL path (e.g., /uploads/avatar/123.jpg)
	return fmt.Sprintf("/%s/%s/%s", constants.DirUploads, config.SubDir, newFileName), nil
}

func GetBaseURL() string {
	// Helper untuk mendapatkan base URL (bisa dari ENV)
	host := os.Getenv("BASE_URL")
	if host == "" {
		return "http://localhost:8080"
	}
	return host
}
