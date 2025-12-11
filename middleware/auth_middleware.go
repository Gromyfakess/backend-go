package middleware

import (
	"strings"

	"siro-backend/repository"
	"siro-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}

		tokenString := strings.Split(authHeader, "Bearer ")[1]

		// 1. Verifikasi Signature JWT
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return utils.JwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(401, gin.H{"error": "Invalid Token Signature"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(401, gin.H{"error": "Invalid Claims"})
			return
		}

		userID := uint(claims["user_id"].(float64))

		// 2. CEK DATABASE (Single Session Enforcer)
		// Memastikan token ini adalah token yang paling baru di DB
		isValidInDB := repository.CheckAccessTokenValid(userID, tokenString)
		if !isValidInDB {
			// Token di header beda dengan di DB (User sudah login di tempat lain / Admin revoke)
			c.AbortWithStatusJSON(401, gin.H{"error": "Session revoked or expired"})
			return
		}

		c.Set("userID", userID)
		c.Set("role", claims["role"].(string))
		c.Set("canCRUD", claims["canCRUD"].(bool))
		c.Next()
	}
}

func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		if role, _ := c.Get("role"); role != "Admin" {
			c.AbortWithStatusJSON(403, gin.H{"error": "Admin only"})
			return
		}
		c.Next()
	}
}
