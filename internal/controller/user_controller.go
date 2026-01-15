package controller

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"siro-backend/global"
	"siro-backend/internal/models"
	"siro-backend/internal/repo"
	"siro-backend/pkg/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

// deleteOldAvatar removes an old avatar file from disk
// Called when a user uploads a new avatar
func deleteOldAvatar(avatarURL string) {
	// Check if URL contains the upload directory path
	targetDir := fmt.Sprintf("/%s/%s/", global.DirUploads, global.DirAvatar)
	if !strings.Contains(avatarURL, targetDir) {
		return // Not a valid avatar URL, skip deletion
	}

	// Extract filename from URL
	parts := strings.Split(avatarURL, "/")
	fileName := parts[len(parts)-1]

	// Security check: prevent directory traversal attacks
	if fileName == "" || fileName == "." || fileName == ".." {
		return
	}

	// Build full file path
	filePath := filepath.Join(global.DirUploads, global.DirAvatar, fileName)

	// Try to delete the file
	// If it fails, log error but don't crash the app
	if err := os.Remove(filePath); err != nil {
		log.Printf("Warning: Failed to delete old avatar file %s: %v", filePath, err)
	}
}

// GetMe returns the current user's information
func GetMe(c *gin.Context) {
	user, ok := getCurrentUser(c)
	if !ok {
		return
	}
	sendSuccess(c, user)
}

// UpdateMe updates the current user's profile
func UpdateMe(c *gin.Context) {
	user, ok := getCurrentUser(c)
	if !ok {
		return
	}

	// Parse request body
	var input models.UserRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		sendError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	// Update user fields
	user.Name = input.Name
	user.Phone = input.Phone

	// Handle avatar update
	if input.AvatarURL != "" {
		// Delete old avatar if user is uploading a new one
		if user.AvatarURL != "" && user.AvatarURL != input.AvatarURL {
			deleteOldAvatar(user.AvatarURL)
		}
		user.AvatarURL = input.AvatarURL
	}

	if input.Password != "" {
		hashedPassword, err := utils.HashPassword(input.Password)
		if err != nil {
			sendError(c, http.StatusInternalServerError, "Failed to hash password")
			return
		}
		user.PasswordHash = hashedPassword
	}

	if err := repo.UpdateUser(user.ID, *user); err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to update profile")
		return
	}

	sendSuccess(c, user)
}

// UploadFile handles file uploads (like avatars)
func UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		sendError(c, http.StatusBadRequest, "File required (max 2MB)")
		return
	}

	uploadConfig := utils.DefaultImageConfig(global.DirAvatar)
	relativePath, err := utils.SaveUploadedFile(file, uploadConfig)
	if err != nil {
		sendError(c, http.StatusBadRequest, err.Error())
		return
	}

	fullURL := utils.GetBaseURL() + relativePath
	sendSuccess(c, gin.H{"url": fullURL})
}

// GetStaffList returns list of staff members based on current user's unit
func GetStaffList(c *gin.Context) {
	user, ok := getCurrentUser(c)
	if !ok {
		return
	}

	staff, err := repo.GetUsersByUnit(user.Unit)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to fetch staff")
		return
	}
	sendSuccess(c, staff)
}

// UpdateAvailability updates a staff member's availability status
func UpdateAvailability(c *gin.Context) {
	userID, ok := parseID(c, "id")
	if !ok {
		return
	}

	var input models.AvailabilityRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		sendError(c, http.StatusBadRequest, "Invalid input")
		return
	}

	validStatuses := map[string]bool{
		global.AvailOnline:  true,
		global.AvailBusy:    true,
		global.AvailAway:    true,
		global.AvailOffline: true,
	}
	if !validStatuses[input.Status] {
		sendError(c, http.StatusBadRequest, "Invalid status. Must be: Online, Busy, Away, or Offline")
		return
	}

	if err := repo.UpdateAvailability(userID, input.Status); err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to update availability")
		return
	}

	sendSuccess(c, gin.H{"message": "Availability updated successfully"})
}

// --- ADMIN ONLY HANDLERS ---

// GetAllUsers returns all users (admin only)
func GetAllUsers(c *gin.Context) {
	users, err := repo.GetAllUsers()
	if err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to fetch users")
		return
	}
	sendSuccess(c, users)
}

// CreateUser creates a new user (admin only)
func CreateUser(c *gin.Context) {
	var input models.UserRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		sendError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	defaultAvatar := fmt.Sprintf("%s/%s/default-avatar.jpg", utils.GetBaseURL(), global.DirUploads)
	if input.AvatarURL != "" {
		defaultAvatar = input.AvatarURL
	}

	newUser := models.User{
		Name:         input.Name,
		Email:        input.Email,
		Role:         input.Role,
		Unit:         input.Unit,
		Phone:        input.Phone,
		CanCRUD:      input.CanCRUD,
		PasswordHash: hashedPassword,
		Availability: global.AvailOffline,
		AvatarURL:    defaultAvatar,
	}

	if err := repo.CreateUser(&newUser); err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to create user")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"statusCode": http.StatusCreated,
		"data":       newUser,
	})
}

// UpdateUser updates an existing user (admin only)
func UpdateUser(c *gin.Context) {
	userID, ok := parseID(c, "id")
	if !ok {
		return
	}

	var input models.UserRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		sendError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	user, err := repo.GetUserByID(userID)
	if err != nil {
		sendError(c, http.StatusNotFound, "User not found")
		return
	}

	// Update user fields
	user.Name = input.Name
	user.Email = input.Email
	user.Role = input.Role
	user.Unit = input.Unit
	user.Phone = input.Phone
	user.CanCRUD = input.CanCRUD

	// Handle avatar update
	if input.AvatarURL != "" {
		// Delete old avatar if user is uploading a new one
		if user.AvatarURL != "" && user.AvatarURL != input.AvatarURL {
			deleteOldAvatar(user.AvatarURL)
		}
		user.AvatarURL = input.AvatarURL
	}

	if input.Password != "" {
		hashedPassword, err := utils.HashPassword(input.Password)
		if err != nil {
			sendError(c, http.StatusInternalServerError, "Failed to hash password")
			return
		}
		user.PasswordHash = hashedPassword
	}

	if err := repo.UpdateUser(user.ID, *user); err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to update user")
		return
	}

	sendSuccess(c, user)
}

// DeleteUser deletes a user (admin only)
func DeleteUser(c *gin.Context) {
	userID, ok := parseID(c, "id")
	if !ok {
		return
	}

	user, err := repo.GetUserByID(userID)
	if err == nil && user.AvatarURL != "" {
		deleteOldAvatar(user.AvatarURL)
	}

	if err := repo.DeleteUser(userID); err != nil {
		sendError(c, http.StatusInternalServerError, "Failed to delete user")
		return
	}

	sendSuccess(c, gin.H{"message": "User deleted successfully"})
}
