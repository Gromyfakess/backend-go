package controllers

import (
	"siro-backend/models"
	"siro-backend/repository"
	"siro-backend/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func LoginHandler(c *gin.Context) {
	var input models.LoginRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	user, err := repository.GetUserByEmail(input.Email)
	if err != nil {
		c.JSON(401, gin.H{"error": "User not found"})
		return
	}

	if err := utils.VerifyPassword(user.PasswordHash, input.Password); err != nil {
		c.JSON(401, gin.H{"error": "Wrong password"})
		return
	}

	// Buat Access & Refresh baru
	acc, ref, atExp, rtExp, err := utils.GenerateAllTokens(user.ID, user.Role, user.CanCRUD)
	if err != nil {
		c.JSON(500, gin.H{"error": "Token gen failed"})
		return
	}

	repository.SaveAllTokens(user.ID, acc, ref, atExp, rtExp)

	// PENTING: Set Cookie (untuk keamanan browser) DAN kirim JSON (untuk NextAuth)
	c.SetCookie("refresh_token", ref, 3600*24*7, "/", "localhost", false, true)

	// Sisa waktu Access Token
	expiresIn := int(time.Until(atExp).Seconds())

	// Kirim token di body
	c.JSON(200, gin.H{
		"accessToken":  acc,
		"refreshToken": ref,
		"expiresIn":    expiresIn,
		"user":         user,
	})
}

func RefreshHandler(c *gin.Context) {
	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(401, gin.H{"error": "No refresh token"})
		return
	}

	// Parse untuk dapat UserID
	token, _ := jwt.Parse(cookie, func(t *jwt.Token) (interface{}, error) {
		return utils.JwtSecret, nil
	})

	if token == nil {
		c.JSON(401, gin.H{"error": "Invalid token"})
		return
	}

	claims, _ := token.Claims.(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64))

	isValid, rtExpiresAt := repository.CheckRefreshTokenValid(userID, cookie)
	if !isValid {
		c.JSON(401, gin.H{"error": "Refresh token invalid/revoked"})
		return
	}

	user, err := repository.GetUserByID(userID)
	if err != nil {
		c.JSON(401, gin.H{"error": "User not found"})
		return
	}

	// --- INTELLIGENT REFRESH LOGIC ---

	timeRemaining := time.Until(rtExpiresAt)
	rotationThreshold := time.Hour * 12

	if timeRemaining < rotationThreshold {
		newAcc, newRef, newAtExp, newRtExp, _ := utils.GenerateAllTokens(user.ID, user.Role, user.CanCRUD)

		repository.SaveAllTokens(user.ID, newAcc, newRef, newAtExp, newRtExp)

		c.SetCookie("refresh_token", newRef, 3600*24*7, "/", "localhost", false, true)
		c.JSON(200, gin.H{"accessToken": newAcc, "status": "rotated"})

	} else {

		newAcc, newAtExp, _ := utils.GenerateAccessTokenOnly(user.ID, user.Role, user.CanCRUD)

		repository.UpdateAccessTokenOnly(user.ID, newAcc, newAtExp)

		c.JSON(200, gin.H{"accessToken": newAcc, "status": "refreshed"})
	}
}

func LogoutHandler(c *gin.Context) {
	c.SetCookie("refresh_token", "", -1, "/", "localhost", false, true)
	c.JSON(200, gin.H{"msg": "Logged out"})
}
