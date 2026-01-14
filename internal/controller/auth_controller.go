package controller

import (
	"log"
	"net/http"
	"siro-backend/internal/models"
	"siro-backend/internal/repo"
	"siro-backend/pkg/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// LoginHandler handles user login
// Returns access token, refresh token, and user info in JSON body
func LoginHandler(c *gin.Context) {
	// Parse login request
	var input models.LoginRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		sendError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	// Find user by email
	user, err := repo.GetUserByEmail(input.Email)
	if err != nil {
		sendError(c, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Verify password
	if err := utils.VerifyPassword(user.PasswordHash, input.Password); err != nil {
		sendError(c, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Generate both tokens (access token + refresh token)
	accessToken, refreshToken, accessExpiry, refreshExpiry, err := utils.GenerateAllTokens(user.ID, user.Role, user.CanCRUD)
	if err != nil {
		log.Printf("Error generating tokens for user %d: %v", user.ID, err)
		sendError(c, http.StatusInternalServerError, "Failed to generate tokens")
		return
	}

	// Save tokens to database (stateful JWT for logout capability)
	err = repo.SaveToken(user.ID, accessToken, refreshToken, accessExpiry, refreshExpiry)
	if err != nil {
		log.Printf("Error saving tokens for user %d: %v", user.ID, err)
		sendError(c, http.StatusInternalServerError, "Failed to save session")
		return
	}

	// Return all tokens and user info in JSON body
	c.JSON(http.StatusOK, gin.H{
		"statusCode":          http.StatusOK,
		"accessToken":         accessToken,
		"accessTokenExpiresAt": accessExpiry.Unix(),
		"refreshToken":        refreshToken,
		"user":                user,
	})
}

// RefreshRequest represents the refresh token request body
type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// RefreshHandler generates a new access token using refresh token
// Refresh token is read from JSON body and stays unchanged
func RefreshHandler(c *gin.Context) {
	// Parse refresh token from request body
	var input RefreshRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		sendError(c, http.StatusBadRequest, "Refresh token is required")
		return
	}

	refreshToken := input.RefreshToken

	// Parse JWT token to get user ID
	token, err := jwt.Parse(refreshToken, func(t *jwt.Token) (interface{}, error) {
		return utils.JwtSecret, nil
	})
	if err != nil {
		sendError(c, http.StatusUnauthorized, "Invalid refresh token")
		return
	}

	// Check if token is valid
	if token == nil || !token.Valid {
		sendError(c, http.StatusUnauthorized, "Invalid refresh token")
		return
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		sendError(c, http.StatusUnauthorized, "Invalid token format")
		return
	}

	// Get user ID from token
	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		sendError(c, http.StatusUnauthorized, "Invalid token: missing user ID")
		return
	}
	userID := uint(userIDFloat)

	// Verify refresh token is still valid in database
	isValid, dbExpiry := repo.CheckRefreshTokenValid(userID, refreshToken)
	if !isValid {
		sendError(c, http.StatusUnauthorized, "Refresh token expired or revoked")
		return
	}

	// Additional check: Verify database expiry
	if !dbExpiry.IsZero() && time.Now().After(dbExpiry) {
		sendError(c, http.StatusUnauthorized, "Refresh token expired")
		return
	}

	// Get latest user data from database
	user, err := repo.GetUserByID(userID)
	if err != nil {
		sendError(c, http.StatusUnauthorized, "User not found")
		return
	}

	// Generate new access token only (refresh token stays the same)
	newAccessToken, newAccessExpiry, err := utils.GenerateAccessTokenOnly(user.ID, user.Role, user.CanCRUD)
	if err != nil {
		log.Printf("Error generating new access token for user %d: %v", user.ID, err)
		sendError(c, http.StatusInternalServerError, "Failed to generate access token")
		return
	}

	// Update only access token in database (refresh token unchanged)
	err = repo.UpdateAccessTokenOnly(user.ID, newAccessToken, newAccessExpiry)
	if err != nil {
		log.Printf("Error updating access token for user %d: %v", user.ID, err)
		sendError(c, http.StatusInternalServerError, "Failed to update session")
		return
	}

	// Return new access token and expiry in JSON body
	c.JSON(http.StatusOK, gin.H{
		"statusCode":          http.StatusOK,
		"accessToken":         newAccessToken,
		"accessTokenExpiresAt": newAccessExpiry.Unix(),
	})
}

// LogoutHandler logs out the current user
// Deletes tokens from database - user must login again after logout
// This is normal behavior: when you logout, tokens are deleted and you must login again
func LogoutHandler(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := getUserID(c)
	if !exists {
		sendError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Delete tokens from database
	// After logout, both access and refresh tokens are deleted
	// User must login again to get new tokens
	if err := repo.DeleteToken(userID); err != nil {
		log.Printf("Warning: Failed to delete token for user %d: %v", userID, err)
		// Continue anyway - logout should succeed even if DB delete fails
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"statusCode": http.StatusOK,
		"message":    "Successfully logged out",
	})
}
