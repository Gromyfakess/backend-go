package utils

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"siro-backend/constants"
	"strings"
	"time"
)

// UploadConfig menyimpan konfigurasi upload dynamic
type UploadConfig struct {
	Folder      string
	AllowedExts map[string]bool
	AllowedMime map[string]bool
	MaxBytes    int64
}

// DefaultImageConfig mengembalikan config standar untuk gambar
func DefaultImageConfig(folder string) UploadConfig {
	return UploadConfig{
		Folder:      folder,
		AllowedExts: map[string]bool{".jpg": true, ".jpeg": true, ".png": true},
		AllowedMime: map[string]bool{"image/jpeg": true, "image/png": true},
		MaxBytes:    constants.MaxFileSize,
	}
}

// SaveUploadedFile menangani validasi, sanitasi, dan penyimpanan file
func SaveUploadedFile(file *multipart.FileHeader, config UploadConfig) (string, error) {
	// 1. Validasi Ukuran
	if file.Size > config.MaxBytes {
		return "", errors.New("file too large (max 2MB)")
	}

	// 2. Validasi Ekstensi
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !config.AllowedExts[ext] {
		return "", errors.New("invalid file extension")
	}

	// 3. Validasi Magic Bytes (Isi file sebenarnya)
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	buffer := make([]byte, 512)
	if _, err := src.Read(buffer); err != nil {
		return "", errors.New("failed to read file header")
	}

	// Reset pointer file kembali ke awal setelah dibaca
	src.Seek(0, 0)

	contentType := http.DetectContentType(buffer)
	if !config.AllowedMime[contentType] {
		return "", errors.New("invalid file content (mime type mismatch)")
	}

	// 4. Persiapan Folder
	uploadDir := filepath.Join("uploads", config.Folder)
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.MkdirAll(uploadDir, 0755)
	}

	// 5. Generate Nama File Unik (Timestamp + Random/Clean Filename)
	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), config.Folder, ext)
	dstPath := filepath.Join(uploadDir, filename)

	// 6. Simpan File (Copy stream)
	out, err := os.Create(dstPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err = io.Copy(out, src); err != nil {
		return "", err
	}

	// 7. Kembalikan URL Relative (Agar frontend yang handle base URL atau Backend handle di response)
	// Menggunakan forward slash agar konsisten di URL
	return fmt.Sprintf("/uploads/%s/%s", config.Folder, filename), nil
}

// GetBaseURL helper untuk generate full URL
func GetBaseURL() string {
	if url := os.Getenv("BACKEND_URL"); url != "" {
		return url
	}
	return "http://localhost:8080"
}
