package utils

import (
	"context"
	"errors"
	"os"
	"time"

	"mime/multipart"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

// UploadToCloudinary: Mengupload file ke Cloudinary dan mengembalikan URL HTTPS
// folderName: nama folder di dalam Cloudinary (misal: "siro/avatars")
func UploadToCloudinary(fileHeader *multipart.FileHeader, folderName string) (string, error) {
	// 1. Ambil URL koneksi dari Environment Variable
	cldURL := os.Getenv("CLOUDINARY_URL")
	if cldURL == "" {
		return "", errors.New("CLOUDINARY_URL is not set")
	}

	// 2. Init Cloudinary
	cld, err := cloudinary.NewFromURL(cldURL)
	if err != nil {
		return "", err
	}

	// 3. Buka file
	src, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// 4. Upload dengan Context Timeout (agar tidak hanging)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Upload params
	uploadParams := uploader.UploadParams{
		Folder: folderName,
	}

	// Eksekusi upload
	resp, err := cld.Upload.Upload(ctx, src, uploadParams)
	if err != nil {
		return "", err
	}

	// 5. Kembalikan URL Secure (HTTPS)
	return resp.SecureURL, nil
}
