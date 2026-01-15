package controller

import (
	"fmt"
	"log"
	"net/http"
	"siro-backend/global"
	"siro-backend/internal/models"
	"siro-backend/internal/repo"
	"siro-backend/pkg/utils"
	"time"

	"github.com/gin-gonic/gin"
)

// GetStats returns dashboard statistics for the current user's unit
func GetStats(c *gin.Context) {
	user, ok := getCurrentUser(c)
	if !ok {
		return
	}

	stats, err := repo.GetDashboardStats(user.Unit)
	if err != nil {
		log.Printf("Error getting stats for unit %s: %v", user.Unit, err)
		sendError(c, http.StatusInternalServerError, "Failed to calculate stats")
		return
	}

	sendSuccess(c, stats)
}

// GetActivities returns paginated activity logs filtered by user's unit
func GetActivities(c *gin.Context) {
	user, ok := getCurrentUser(c)
	if !ok {
		return
	}

	pagination := getPaginationParams(c)
	logs, meta, err := repo.GetActivities(user.Unit, pagination.Page, pagination.Limit)
	if err != nil {
		log.Printf("Error getting activities: %v", err)
		sendError(c, http.StatusInternalServerError, "Failed to fetch activities")
		return
	}

	sendPaginatedResponse(c, logs, meta)
}

// CreateWorkOrder creates a new request
func CreateWorkOrder(c *gin.Context) {
	user, ok := getCurrentUser(c)
	if !ok {
		return
	}

	// Check permissions
	role, _ := c.Get("role")
	canCRUD, _ := c.Get("canCRUD")
	if role != global.RoleAdmin && !canCRUD.(bool) {
		sendError(c, http.StatusForbidden, "Permission denied")
		return
	}

	// Parse request body
	var input models.WorkOrderRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		sendError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	// Validate unit
	if input.Unit == "" {
		sendError(c, http.StatusBadRequest, "Unit must be selected")
		return
	}

	// SECURITY CHECK: Cannot create request for own unit
	if input.Unit == user.Unit {
		sendError(c, http.StatusBadRequest, "You cannot create a request for your own unit")
		return
	}

	// Create request
	newOrder := models.WorkOrder{
		Title:       input.Title,
		Description: input.Description,
		Priority:    input.Priority,
		RequesterID: user.ID,
		Unit:        input.Unit,
		PhotoURL:    input.PhotoURL,
		Status:      global.StatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := repo.CreateWorkOrder(&newOrder); err != nil {
		log.Printf("Error creating request: %v", err)
		sendError(c, http.StatusInternalServerError, "Failed to create request")
		return
	}

	// Get full request details
	fullOrder, err := repo.GetWorkOrderById(newOrder.ID)
	if err != nil {
		log.Printf("Error retrieving created request %d: %v", newOrder.ID, err)
		sendError(c, http.StatusInternalServerError, "Request created but failed to retrieve details")
		return
	}

	// Log activity
	repo.LogActivity(user.ID, user.Name, fmt.Sprintf("created request to %s:", input.Unit), fullOrder.Title, global.StatusPending, fullOrder.ID)

	c.JSON(http.StatusCreated, gin.H{
		"statusCode": http.StatusCreated,
		"data":       fullOrder,
	})
}

// UploadWorkOrderEvidence handles file upload for request evidence
func UploadWorkOrderEvidence(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		sendError(c, http.StatusBadRequest, "No file uploaded")
		return
	}

	uploadConfig := utils.DefaultImageConfig(global.DirWorkOrder)
	relativePath, err := utils.SaveUploadedFile(file, uploadConfig)
	if err != nil {
		sendError(c, http.StatusBadRequest, err.Error())
		return
	}

	fullURL := utils.GetBaseURL() + relativePath
	c.JSON(http.StatusOK, gin.H{
		"statusCode": http.StatusOK,
		"message":    "File uploaded successfully",
		"url":        fullURL,
	})
}

// GetWorkOrders returns paginated list of requests with filters
func GetWorkOrders(c *gin.Context) {
	pagination := getPaginationParams(c)

	filters := map[string]string{
		"status":         c.Query("status"),
		"unit":           c.Query("unit"),
		"requester_unit": c.Query("requester_unit"),
		"date":           c.Query("date"),
	}

	orders, meta, err := repo.GetWorkOrders(filters, pagination.Page, pagination.Limit)
	if err != nil {
		log.Printf("Error getting requests: %v", err)
		sendError(c, http.StatusInternalServerError, "Failed to fetch requests")
		return
	}

	sendPaginatedResponse(c, orders, meta)
}

// TakeRequest allows a staff member to take/claim a request
func TakeRequest(c *gin.Context) {
	user, ok := getCurrentUser(c)
	if !ok {
		return
	}

	orderID, ok := parseID(c, "id")
	if !ok {
		return
	}

	order, err := repo.GetWorkOrderById(orderID)
	if err != nil {
		sendError(c, http.StatusNotFound, "Request not found")
		return
	}

	// SECURITY CHECK: You can only take a request if it is assigned to YOUR unit.
	if order.Unit != user.Unit {
		sendError(c, http.StatusForbidden, "You cannot take a request assigned to another unit.")
		return
	}

	if order.Status == global.StatusCompleted {
		sendError(c, http.StatusBadRequest, "Request already completed")
		return
	}

	if err := repo.TakeWorkOrder(orderID, user.ID); err != nil {
		sendError(c, http.StatusConflict, "Failed to take request. It may have been taken by someone else.")
		return
	}

	repo.LogActivity(user.ID, user.Name, "is working on:", order.Title, global.StatusInProgress, order.ID)
	sendSuccess(c, gin.H{"message": "Request taken successfully"})
}

// AssignStaff allows admin to assign a request to a staff member
func AssignStaff(c *gin.Context) {
	admin, ok := getCurrentUser(c)
	if !ok {
		return
	}

	orderID, ok := parseID(c, "id")
	if !ok {
		return
	}

	var input models.AssignRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		sendError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	// Fetch Order First to check permissions
	order, err := repo.GetWorkOrderById(orderID)
	if err != nil {
		sendError(c, http.StatusNotFound, "Request not found")
		return
	}

	// SECURITY CHECK: Only users from the target unit can assign staff.
	if order.Unit != admin.Unit {
		sendError(c, http.StatusForbidden, "You cannot assign staff to a request for another unit.")
		return
	}

	// Verify Assignee
	assignee, err := repo.GetUserByID(input.AssigneeID)
	if err != nil {
		sendError(c, http.StatusNotFound, "Staff member not found")
		return
	}

	// SECURITY CHECK: Assignee must be from the same unit
	if assignee.Unit != admin.Unit {
		sendError(c, http.StatusBadRequest, "Assignee must be from the same unit")
		return
	}

	if order.Status == global.StatusCompleted {
		sendError(c, http.StatusBadRequest, "Request already completed")
		return
	}

	if err := repo.AssignWorkOrder(orderID, input.AssigneeID); err != nil {
		log.Printf("Error assigning request %d to user %d: %v", orderID, input.AssigneeID, err)
		sendError(c, http.StatusInternalServerError, "Failed to assign staff")
		return
	}

	repo.LogActivity(admin.ID, admin.Name, fmt.Sprintf("assigned request to %s:", assignee.Name), order.Title, global.StatusInProgress, order.ID)
	sendSuccess(c, gin.H{"message": "Staff assigned successfully"})
}

// FinalizeOrder marks a request as completed
func FinalizeOrder(c *gin.Context) {
	user, ok := getCurrentUser(c)
	if !ok {
		return
	}

	orderID, ok := parseID(c, "id")
	if !ok {
		return
	}

	var input models.FinalizeRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		input.Note = "" // Note is optional
	}

	order, err := repo.GetWorkOrderById(orderID)
	if err != nil {
		sendError(c, http.StatusNotFound, "Request not found")
		return
	}

	// SECURITY CHECK: Only the unit responsible for the work can finalize it.
	if order.Unit != user.Unit {
		sendError(c, http.StatusForbidden, "You cannot finalize a request belonging to another unit.")
		return
	}

	if order.AssigneeID == nil {
		sendError(c, http.StatusBadRequest, "Request has not been assigned yet")
		return
	}

	// Check permission: only assignee or admin can finalize
	role, _ := c.Get("role")
	if *order.AssigneeID != user.ID && role != global.RoleAdmin {
		sendError(c, http.StatusForbidden, "Access denied")
		return
	}

	if err := repo.FinalizeWorkOrder(orderID, input.Note, user.ID); err != nil {
		log.Printf("Error finalizing request %d: %v", orderID, err)
		sendError(c, http.StatusInternalServerError, "Failed to finalize request")
		return
	}

	repo.LogActivity(user.ID, user.Name, "completed request:", order.Title, global.StatusCompleted, order.ID)
	sendSuccess(c, gin.H{"message": "Request finalized successfully"})
}
