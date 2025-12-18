package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"siro-backend/models"
	"siro-backend/repository"
	"siro-backend/utils"

	"github.com/gin-gonic/gin"
)

// --- HELPER FUNCTION ---

// deleteOldAvatar: Menghapus file foto lama dari server
func deleteOldAvatar(avatarURL string) {
	// Pastikan URL mengandung path avatar yang benar
	if !strings.Contains(avatarURL, "/uploads/avatar/") {
		return
	}

	// Jangan hapus default avatar
	if strings.Contains(avatarURL, "default-avatar") {
		return
	}

	// Ambil nama file dari URL
	parts := strings.Split(avatarURL, "/")
	fileName := parts[len(parts)-1]

	// Hapus file fisik di folder uploads/avatar
	filePath := filepath.Join("uploads", "avatar", fileName)
	err := os.Remove(filePath)
	if err != nil {
		fmt.Println("Warning: Gagal menghapus foto lama:", fileName, err)
	} else {
		fmt.Println("Berhasil menghapus foto lama:", fileName)
	}
}

// --- PUBLIC / AUTHENTICATED USER HANDLERS ---

func GetMe(c *gin.Context) {
	uid, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	user, err := repository.GetUserByID(uid.(uint))
	if err != nil {
		c.JSON(404, gin.H{"error": "User not found"})
		return
	}
	c.JSON(200, user)
}

func UpdateMe(c *gin.Context) {
	uid, _ := c.Get("userID")

	user, err := repository.GetUserByID(uid.(uint))
	if err != nil {
		c.JSON(404, gin.H{"error": "User not found"})
		return
	}

	var i models.UserRequest
	if err := c.ShouldBindJSON(&i); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	user.Name = i.Name
	user.Phone = i.Phone

	if i.AvatarURL != "" {
		if user.AvatarURL != "" && user.AvatarURL != i.AvatarURL {
			deleteOldAvatar(user.AvatarURL)
		}
		user.AvatarURL = i.AvatarURL
	}

	if i.Password != "" {
		p, _ := utils.HashPassword(i.Password)
		user.PasswordHash = p
	}

	if err := repository.UpdateUser(&user); err != nil {
		c.JSON(500, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(200, user)
}

// UploadFile: Upload gambar avatar ke folder uploads/avatar
func UploadFile(c *gin.Context) {
	// Batasan Ukuran File (Max 2MB)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 2<<20)

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "File required (max 2MB)"})
		return
	}

	// Validasi Ekstensi
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true}
	if !allowedExts[ext] {
		c.JSON(400, gin.H{"error": "File type not allowed. Only .jpg, .jpeg, .png"})
		return
	}

	// Validasi Konten (Magic Bytes)
	src, err := file.Open()
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to open file"})
		return
	}
	defer src.Close()

	buffer := make([]byte, 512)
	_, err = src.Read(buffer)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to sniff file"})
		return
	}
	src.Seek(0, 0) // Reset pointer

	contentType := http.DetectContentType(buffer)
	allowedMimes := map[string]bool{"image/jpeg": true, "image/png": true}
	if !allowedMimes[contentType] {
		c.JSON(400, gin.H{"error": "Invalid file content"})
		return
	}

	// --- LOGIKA SAVE PATH BARU ---

	// 1. Buat nama file unik
	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), "avatar", ext)

	// 2. Tentukan folder tujuan: uploads/avatar
	uploadDir := filepath.Join("uploads", "avatar")

	// 3. Buat folder jika belum ada
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.MkdirAll(uploadDir, 0755) // MkdirAll bisa buat nested folder (uploads/avatar)
	}

	// 4. Path lengkap file
	savePath := filepath.Join(uploadDir, filename)

	// 5. Simpan File
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(500, gin.H{"error": "Failed to save file"})
		return
	}

	// 6. Generate URL untuk Frontend
	backendURL := "http://localhost:8080"
	if envURL := os.Getenv("BACKEND_URL"); envURL != "" {
		backendURL = envURL
	}

	// URL harus menggunakan slash (/), bukan backslash (\)
	// Output: http://localhost:8080/uploads/avatar/namafile.jpg
	fullURL := fmt.Sprintf("%s/uploads/avatar/%s", backendURL, filename)

	c.JSON(200, gin.H{"url": fullURL})
}

func GetStaffList(c *gin.Context) {
	users := repository.GetStaffByUnit("IT Center")
	c.JSON(200, users)
}

func UpdateAvailability(c *gin.Context) {
	id := c.Param("id")
	var i models.AvailabilityRequest
	if err := c.ShouldBindJSON(&i); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	if err := repository.UpdateAvailability(id, i.Status); err != nil {
		c.JSON(500, gin.H{"error": "Failed update availability"})
		return
	}
	c.JSON(200, gin.H{"status": "ok"})
}

// --- ADMIN HANDLERS ---

func GetAllUsers(c *gin.Context) {
	users := repository.GetAllUsers()
	c.JSON(200, users)
}

func CreateUser(c *gin.Context) {
	var i models.UserRequest
	if err := c.ShouldBindJSON(&i); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	p, _ := utils.HashPassword(i.Password)

	backendURL := "http://localhost:8080"
	if envURL := os.Getenv("BACKEND_URL"); envURL != "" {
		backendURL = envURL
	}

	defaultAvatar := fmt.Sprintf("%s/uploads/default-avatar.jpg", backendURL)

	if i.AvatarURL != "" {
		defaultAvatar = i.AvatarURL
	}

	u := models.User{
		Name:         i.Name,
		Email:        i.Email,
		Role:         i.Role,
		Unit:         i.Unit,
		Phone:        i.Phone,
		CanCRUD:      i.CanCRUD,
		PasswordHash: p,
		Availability: "Offline",
		AvatarURL:    defaultAvatar,
	}

	if err := repository.CreateUser(&u); err != nil {
		c.JSON(500, gin.H{"error": "Failed to create user"})
		return
	}
	c.JSON(201, u)
}

func UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	var i models.UserRequest
	if err := c.ShouldBindJSON(&i); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	id, _ := strconv.Atoi(idStr)
	user, err := repository.GetUserByID(uint(id))
	if err != nil {
		c.JSON(404, gin.H{"error": "User not found"})
		return
	}

	user.Name = i.Name
	user.Email = i.Email
	user.Role = i.Role
	user.Unit = i.Unit
	user.Phone = i.Phone
	user.CanCRUD = i.CanCRUD

	if i.AvatarURL != "" {
		if user.AvatarURL != "" && user.AvatarURL != i.AvatarURL {
			deleteOldAvatar(user.AvatarURL)
		}
		user.AvatarURL = i.AvatarURL
	}

	if i.Password != "" {
		p, _ := utils.HashPassword(i.Password)
		user.PasswordHash = p
	}

	if err := repository.UpdateUser(&user); err != nil {
		c.JSON(500, gin.H{"error": "Failed to update user"})
		return
	}
	c.JSON(200, user)
}

func DeleteUser(c *gin.Context) {
	idStr := c.Param("id")

	id, _ := strconv.Atoi(idStr)
	user, err := repository.GetUserByID(uint(id))
	if err == nil {
		deleteOldAvatar(user.AvatarURL)
	}

	if err := repository.DeleteUser(idStr); err != nil {
		c.JSON(500, gin.H{"error": "Failed to delete user"})
		return
	}
	c.JSON(200, gin.H{"message": "Deleted"})
}
