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

// --- PUBLIC / AUTHENTICATED USER HANDLERS ---

// GetMe: Mengambil data user yang sedang login
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

// UpdateMe: User mengupdate datanya sendiri (Nama, Telepon, Avatar)
func UpdateMe(c *gin.Context) {
	uid, _ := c.Get("userID") // Ambil ID dari token

	// Ambil data user lama
	user, err := repository.GetUserByID(uid.(uint))
	if err != nil {
		c.JSON(404, gin.H{"error": "User not found"})
		return
	}

	// Bind input JSON baru
	var i models.UserRequest
	if err := c.ShouldBindJSON(&i); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Update field
	user.Name = i.Name
	user.Phone = i.Phone

	// Update avatar jika dikirim string URL-nya
	if i.AvatarURL != "" {
		user.AvatarURL = i.AvatarURL
	}

	// Update password opsional (hanya jika diisi)
	if i.Password != "" {
		p, _ := utils.HashPassword(i.Password)
		user.PasswordHash = p
	}

	// 4. Simpan ke database
	if err := repository.UpdateUser(&user); err != nil {
		c.JSON(500, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(200, user)
}

// UploadFile: Upload gambar avatar dengan keamanan ketat (Anti-Webshell)
func UploadFile(c *gin.Context) {
	// 1. Batasi Ukuran File (Max 2MB)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 2<<20)

	// Ambil file dari form-data
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

	// Validasi Konten File (Magic Bytes / MIME Sniffing)
	src, err := file.Open()
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to open file"})
		return
	}
	defer src.Close()

	// Baca 512 byte pertama untuk deteksi tipe konten asli
	buffer := make([]byte, 512)
	_, err = src.Read(buffer)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to sniff file"})
		return
	}

	// Reset pointer file kembali ke awal agar bisa disimpan utuh nanti
	src.Seek(0, 0)

	contentType := http.DetectContentType(buffer)
	allowedMimes := map[string]bool{"image/jpeg": true, "image/png": true}
	if !allowedMimes[contentType] {
		c.JSON(400, gin.H{"error": "Invalid file content (fake extension detected)"})
		return
	}

	// Sanitasi & Rename Nama File (Mencegah Path Traversal & Overwrite)
	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), "avatar", ext)
	savePath := filepath.Join("uploads", filename)

	// Pastikan folder uploads ada
	if _, err := os.Stat("uploads"); os.IsNotExist(err) {
		os.Mkdir("uploads", 0755)
	}

	// Simpan File ke Disk
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(500, gin.H{"error": "Failed to save file"})
		return
	}

	backendURL := "http://localhost:8080"

	if envURL := os.Getenv("BACKEND_URL"); envURL != "" {
		backendURL = envURL
	}

	// URL Static File
	fullURL := fmt.Sprintf("%s/uploads/%s", backendURL, filename)

	c.JSON(200, gin.H{"url": fullURL})
}

// GetStaffList: Mengambil list staff IT Center
func GetStaffList(c *gin.Context) {
	users := repository.GetStaffByUnit("IT Center")
	c.JSON(200, users)
}

// UpdateAvailability: Update status online/busy/away
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

// --- ADMIN HANDLERS (CRUD Users) ---

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

	// Hash password sebelum disimpan
	p, _ := utils.HashPassword(i.Password)

	baseURL := os.Getenv("BACKEND_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	defaultAvatar := fmt.Sprintf("%s/uploads/default-avatar.jpg", baseURL)

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

	// Admin berhak update semua field
	user.Name = i.Name
	user.Email = i.Email
	user.Role = i.Role
	user.Unit = i.Unit
	user.Phone = i.Phone // <--- Update Phone
	user.CanCRUD = i.CanCRUD

	// Update password hanya jika admin mengisi field password
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
	id := c.Param("id")
	if err := repository.DeleteUser(id); err != nil {
		c.JSON(500, gin.H{"error": "Failed to delete user"})
		return
	}
	c.JSON(200, gin.H{"message": "Deleted"})
}
