package middleware

import (
	"siro-backend/constants"
	"siro-backend/repository"
	"siro-backend/utils"
	"strings"

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

		isValidInDB := repository.CheckAccessTokenValid(userID, tokenString)
		if !isValidInDB {
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
		role, exists := c.Get("role")
		if !exists || role != constants.RoleAdmin {
			c.AbortWithStatusJSON(403, gin.H{"error": "Admin only area"})
			return
		}
		c.Next()
	}
}
