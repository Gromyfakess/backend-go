package controllers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"siro-backend/models"
	"siro-backend/repository"
	"siro-backend/utils"

	"github.com/gin-gonic/gin"
)

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

	// Update Avatar URL (URL didapat dari Cloudinary di frontend atau via upload endpoint terpisah)
	if i.AvatarURL != "" {
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

// UploadFile: Upload gambar avatar ke Cloudinary
// Endpoint: POST /upload
func UploadFile(c *gin.Context) {
	// 1. Batasan Ukuran File (Max 2MB)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 2<<20)

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "File required (max 2MB)"})
		return
	}

	// 2. Validasi Ekstensi
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true}
	if !allowedExts[ext] {
		c.JSON(400, gin.H{"error": "File type not allowed. Only .jpg, .jpeg, .png"})
		return
	}

	// 3. UPLOAD KE CLOUDINARY
	// Gunakan helper yang sudah dibuat
	imgURL, err := utils.UploadToCloudinary(file, "siro/avatars")
	if err != nil {
		fmt.Println("Cloudinary Error:", err)
		c.JSON(500, gin.H{"error": "Gagal mengupload ke Cloud Storage"})
		return
	}

	// 4. Return URL Cloudinary
	c.JSON(200, gin.H{"url": imgURL})
}

func GetStaffList(c *gin.Context) {
	users := repository.GetStaffByUnit("IT Center")
	c.JSON(200, users)
}

func UpdateAvailability(c *gin.Context) {
	idStr := c.Param("id")
	// Konversi string ID ke uint jika perlu di repository, tapi biasanya param di Gin string
	// Asumsi repository.UpdateAvailability menerima ID string atau uint (sesuaikan)

	var i models.AvailabilityRequest
	if err := c.ShouldBindJSON(&i); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	if err := repository.UpdateAvailability(idStr, i.Status); err != nil {
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

	// Default Avatar Cloudinary (Pastikan Anda punya file ini di Cloudinary atau gunakan placeholder umum)
	// Jika tidak, bisa pakai link statis umum
	defaultAvatar := "https://res.cloudinary.com/demo/image/upload/v1/user.png"

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
	if err := repository.DeleteUser(idStr); err != nil {
		c.JSON(500, gin.H{"error": "Failed to delete user"})
		return
	}
	c.JSON(200, gin.H{"message": "Deleted"})
}
