package controllers

import (
	"fmt"
	"os"
	"path/filepath"
	"siro-backend/constants"
	"siro-backend/models"
	"siro-backend/repository"
	"siro-backend/utils"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// --- HELPER FUNCTION ---

// deleteOldAvatar: Menghapus file foto lama dari server
func deleteOldAvatar(avatarURL string) {
	targetDir := fmt.Sprintf("/uploads/%s/", "avatar")
	if !strings.Contains(avatarURL, targetDir) {
		return
	}

	parts := strings.Split(avatarURL, "/")
	fileName := parts[len(parts)-1]

	filePath := filepath.Join("uploads", "avatar", fileName)
	_ = os.Remove(filePath)
}

// --- HANDLERS ---

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

func UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "File required (max 2MB)"})
		return
	}

	uploadConfig := utils.DefaultImageConfig("avatar")
	relativePath, err := utils.SaveUploadedFile(file, uploadConfig)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	fullURL := utils.GetBaseURL() + relativePath
	c.JSON(200, gin.H{"url": fullURL})
}

func GetStaffList(c *gin.Context) {
	users := repository.GetStaffByUnit(constants.DefaultUnit)
	c.JSON(200, users)
}

func UpdateAvailability(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID format"})
		return
	}

	var i models.AvailabilityRequest
	if err := c.ShouldBindJSON(&i); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	validStatuses := map[string]bool{
		constants.AvailOnline:  true,
		constants.AvailBusy:    true,
		constants.AvailAway:    true,
		constants.AvailOffline: true,
	}
	if !validStatuses[i.Status] {
		c.JSON(400, gin.H{"error": "Invalid status"})
		return
	}

	if err := repository.UpdateAvailability(uint(id), i.Status); err != nil {
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
	defaultAvatar := fmt.Sprintf("%s/uploads/default-avatar.jpg", utils.GetBaseURL())

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
		Availability: constants.AvailOffline,
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

	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID format"})
		return
	}

	var i models.UserRequest
	if err := c.ShouldBindJSON(&i); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

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

	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID format"})
		return
	}

	user, err := repository.GetUserByID(uint(id))
	if err == nil {
		deleteOldAvatar(user.AvatarURL)
	}

	if err := repository.DeleteUser(uint(id)); err != nil {
		c.JSON(500, gin.H{"error": "Failed to delete user"})
		return
	}
	c.JSON(200, gin.H{"message": "Deleted"})
}
