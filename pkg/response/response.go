package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Success sends a successful JSON response
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

// Error sends an error JSON response
func Error(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{
		"success": false,
		"error":   message,
	})
}

// Paginated sends a paginated JSON response
func Paginated(c *gin.Context, data interface{}, meta interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
		"meta":    meta,
	})
}
