package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"siro-backend/constants" // Import constants
	"siro-backend/models"
	"siro-backend/repository"
	"siro-backend/utils"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func deleteOldAvatar(avatarURL string) {
	// FIX: Gunakan constants
	targetDir := fmt.Sprintf("/%s/%s/", constants.DirUploads, constants.DirAvatar)

	if !strings.Contains(avatarURL, targetDir) {
		return
	}
	parts := strings.Split(avatarURL, "/")
	fileName := parts[len(parts)-1]
	if fileName == "" || fileName == "." || fileName == ".." {
		return
	}

	// FIX: Gunakan constants untuk path OS
	filePath := filepath.Join(constants.DirUploads, constants.DirAvatar, fileName)
	_ = os.Remove(filePath)
}

func GetMe(c *gin.Context) {
	uid, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	user, err := repository.GetUserByID(uid.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func UpdateMe(c *gin.Context) {
	uid, _ := c.Get("userID")
	userID := uid.(uint)

	user, err := repository.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var i models.UserRequest
	if err := c.ShouldBindJSON(&i); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

	if err := repository.UpdateUser(user.ID, *user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File required (max 2MB)"})
		return
	}

	// FIX: Gunakan constants
	uploadConfig := utils.DefaultImageConfig(constants.DirAvatar)

	relativePath, err := utils.SaveUploadedFile(file, uploadConfig)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fullURL := utils.GetBaseURL() + relativePath
	c.JSON(http.StatusOK, gin.H{"url": fullURL})
}

func GetStaffList(c *gin.Context) {
	staff, err := repository.GetUsersByUnit(constants.DefaultUnit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch staff"})
		return
	}
	c.JSON(http.StatusOK, staff)
}

func UpdateAvailability(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var i models.AvailabilityRequest
	if err := c.ShouldBindJSON(&i); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	validStatuses := map[string]bool{
		constants.AvailOnline:  true,
		constants.AvailBusy:    true,
		constants.AvailAway:    true,
		constants.AvailOffline: true,
	}
	if !validStatuses[i.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status"})
		return
	}

	if err := repository.UpdateAvailability(uint(id), i.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed update availability"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// --- ADMIN HANDLERS ---

func GetAllUsers(c *gin.Context) {
	users, err := repository.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}
	c.JSON(http.StatusOK, users)
}

func CreateUser(c *gin.Context) {
	var i models.UserRequest
	if err := c.ShouldBindJSON(&i); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	p, _ := utils.HashPassword(i.Password)

	// FIX: Gunakan constants untuk default avatar path
	defaultAvatar := fmt.Sprintf("%s/%s/default-avatar.jpg", utils.GetBaseURL(), constants.DirUploads)

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}
	c.JSON(http.StatusCreated, u)
}

func UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var i models.UserRequest
	if err := c.ShouldBindJSON(&i); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := repository.GetUserByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
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

	if err := repository.UpdateUser(user.ID, *user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	user, err := repository.GetUserByID(uint(id))
	if err == nil && user.AvatarURL != "" {
		deleteOldAvatar(user.AvatarURL)
	}

	if err := repository.DeleteUser(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
}
