package middlewares

import (
	"fmt"
	"net/http"
	"siro-backend/global"
	"siro-backend/internal/repo"
	"siro-backend/pkg/utils"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// --- BYPASS OPTIONS (PREFLIGHT) ---
		// Jika method OPTIONS, langsung return 204 No Content.
		// Jangan divalidasi tokennya, karena browser tidak kirim token saat preflight.
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		// 1. Ambil Header Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		// 2. Format harus "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			return
		}

		tokenString := parts[1]

		// 3. Parse Token
		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return utils.JwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		// 4. Ekstrak Claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// Handle user_id
			if idFloat, ok := claims["user_id"].(float64); ok {
				c.Set("userID", uint(idFloat))
			} else {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims: user_id"})
				return
			}

			// Handle role
			if role, ok := claims["role"].(string); ok {
				c.Set("role", role)
			} else {
				c.Set("role", "")
			}

			// Handle canCRUD (Sesuai dengan key di token.go)
			if canCRUD, ok := claims["canCRUD"].(bool); ok {
				c.Set("canCRUD", canCRUD)
			} else {
				c.Set("canCRUD", false)
			}

			// Validasi Database (Strict)
			userID := uint(claims["user_id"].(float64))
			if !repo.CheckAccessTokenValid(userID, tokenString) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Session expired or logged out"})
				return
			}

		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}

		c.Next()
	}
}

func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Bypass OPTIONS juga untuk Admin route
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		role, exists := c.Get("role")
		if !exists || role != global.RoleAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden: Admins only"})
			return
		}
		c.Next()
	}
}
