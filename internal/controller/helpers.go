package controller

import (
	"net/http"
	"siro-backend/internal/models"
	"siro-backend/internal/repo"
	"strconv"

	"github.com/gin-gonic/gin"
)

// getUserID extracts user ID from context (set by auth middleware)
// Returns 0 and false if not found
func getUserID(c *gin.Context) (uint, bool) {
	uid, exists := c.Get("userID")
	if !exists {
		return 0, false
	}
	userID, ok := uid.(uint)
	return userID, ok
}

// getCurrentUser retrieves the current authenticated user from database
// Returns error response if user not found
func getCurrentUser(c *gin.Context) (*models.User, bool) {
	userID, exists := getUserID(c)
	if !exists {
		sendError(c, http.StatusUnauthorized, "Unauthorized")
		return nil, false
	}

	user, err := repo.GetUserByID(userID)
	if err != nil {
		sendError(c, http.StatusUnauthorized, "User not found")
		return nil, false
	}

	return user, true
}

// parseID extracts and validates ID from URL parameter
func parseID(c *gin.Context, paramName string) (uint, bool) {
	idStr := c.Param(paramName)
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 1 {
		sendError(c, http.StatusBadRequest, "Invalid ID format")
		return 0, false
	}
	return uint(id), true
}

// PaginationParams represents pagination parameters
type PaginationParams struct {
	Page  int
	Limit int
}

// getPaginationParams extracts and validates pagination parameters from query string
// Defaults: page=1, limit=10, max limit=100
func getPaginationParams(c *gin.Context) PaginationParams {
	// Parse page (default: 1)
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	// Parse limit (default: 10, max: 100)
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	return PaginationParams{Page: page, Limit: limit}
}

// sendPaginatedResponse sends a paginated response with status code
func sendPaginatedResponse(c *gin.Context, data interface{}, meta models.PaginationMeta) {
	// Ensure data is never null
	if data == nil {
		data = []interface{}{}
	}
	c.JSON(http.StatusOK, gin.H{
		"statusCode": http.StatusOK,
		"data":       data,
		"meta":       meta,
	})
}

// sendError sends an error response with status code in body
func sendError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{
		"statusCode": statusCode,
		"error":      message,
	})
}

// sendSuccess sends a success response with status code in body
func sendSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"statusCode": http.StatusOK,
		"data":       data,
	})
}
