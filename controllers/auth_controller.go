package controllers

import (
	"net/http"
	"os"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// 1. Cek User via Repository
	user, err := repository.GetUserByEmail(input.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// 2. Verifikasi Password
	if err := utils.VerifyPassword(user.PasswordHash, input.Password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Wrong password"})
		return
	}

	// 3. Generate Token Baru
	acc, ref, atExp, rtExp, err := utils.GenerateAllTokens(user.ID, user.Role, user.CanCRUD)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
		return
	}

	// 4. Simpan Token ke Database (Menggunakan repository.SaveToken)
	err = repository.SaveToken(user.ID, acc, ref, atExp, rtExp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}

	// 5. Set Cookie Refresh Token
	isProduction := os.Getenv("APP_ENV") == "production"
	c.SetCookie("refresh_token", ref, 3600*24*7, "/", "", isProduction, true)

	// Hitung sisa waktu Access Token untuk frontend
	expiresIn := int(time.Until(atExp).Seconds())

	c.JSON(http.StatusOK, gin.H{
		"accessToken":  acc,
		"refreshToken": ref,
		"expiresIn":    expiresIn,
		"user":         user,
	})
}

func RefreshHandler(c *gin.Context) {
	// 1. Ambil Cookie
	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No refresh token provided"})
		return
	}

	// 2. Parse JWT Tanpa Verifikasi Signature dulu (hanya untuk ambil UserID)
	// Kita akan validasi keamanan via Database Check di langkah selanjutnya
	token, _ := jwt.Parse(cookie, func(t *jwt.Token) (interface{}, error) {
		return utils.JwtSecret, nil
	})

	// Pastikan struktur token valid
	if token == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token structure"})
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		return
	}

	// Konversi UserID dari float64 (standar JSON JWT) ke uint
	userID := uint(claims["user_id"].(float64))

	// 3. Validasi Refresh Token ke Database
	// (Ini memastikan token belum dilogout atau direplace user lain)
	isValid, _ := repository.CheckRefreshTokenValid(userID, cookie)
	if !isValid {
		// Hapus cookie karena sudah tidak valid
		c.SetCookie("refresh_token", "", -1, "/", "", false, true)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token expired or revoked"})
		return
	}

	// 4. Ambil Data User Terbaru
	user, err := repository.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// 5. Generate HANYA Access Token Baru (Refresh Token lama tetap dipakai sampai expire)
	newAcc, newAtExp, err := utils.GenerateAccessTokenOnly(user.ID, user.Role, user.CanCRUD)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	// 6. Update Database (Hanya kolom access_token)
	err = repository.UpdateAccessTokenOnly(user.ID, newAcc, newAtExp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update session"})
		return
	}

	// 7. Response
	expiresIn := int(time.Until(newAtExp).Seconds())
	c.JSON(http.StatusOK, gin.H{
		"accessToken": newAcc,
		"expiresIn":   expiresIn,
		"status":      "refreshed",
	})
}

func LogoutHandler(c *gin.Context) {
	// 1. Coba ambil UserID dari cookie untuk menghapus sesi di Database
	cookie, err := c.Cookie("refresh_token")
	if err == nil && cookie != "" {
		token, _ := jwt.Parse(cookie, func(t *jwt.Token) (interface{}, error) {
			return utils.JwtSecret, nil
		})

		if token != nil {
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				// Ambil UserID
				if idFloat, ok := claims["user_id"].(float64); ok {
					userID := uint(idFloat)
					_ = repository.DeleteToken(userID)
				}
			}
		}
	}

	isProduction := os.Getenv("APP_ENV") == "production"
	c.SetCookie("refresh_token", "", -1, "/", "", isProduction, true)

	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}
