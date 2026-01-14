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

// CreateWorkOrder creates a new work order/ticket
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

	// Create work order
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
		log.Printf("Error creating work order: %v", err)
		sendError(c, http.StatusInternalServerError, "Failed to create work order")
		return
	}

	// Get full order details
	fullOrder, err := repo.GetWorkOrderById(newOrder.ID)
	if err != nil {
		log.Printf("Error retrieving created work order %d: %v", newOrder.ID, err)
		sendError(c, http.StatusInternalServerError, "Work order created but failed to retrieve details")
		return
	}

	// Log activity
	repo.LogActivity(user.ID, user.Name, fmt.Sprintf("created request to %s:", input.Unit), fullOrder.Title, global.StatusPending, fullOrder.ID)

	c.JSON(http.StatusCreated, gin.H{
		"statusCode": http.StatusCreated,
		"data":       fullOrder,
	})
}

// UploadWorkOrderEvidence handles file upload for work order evidence
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

// GetWorkOrders returns paginated list of work orders with filters
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
		log.Printf("Error getting work orders: %v", err)
		sendError(c, http.StatusInternalServerError, "Failed to fetch work orders")
		return
	}

	sendPaginatedResponse(c, orders, meta)
}

// TakeRequest allows a staff member to take/claim a work order
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
		sendError(c, http.StatusNotFound, "Work order not found")
		return
	}

	if order.Status == global.StatusCompleted {
		sendError(c, http.StatusBadRequest, "Work order already completed")
		return
	}

	if err := repo.TakeWorkOrder(orderID, user.ID); err != nil {
		sendError(c, http.StatusConflict, "Failed to take work order. It may have been taken by someone else.")
		return
	}

	repo.LogActivity(user.ID, user.Name, "is working on:", order.Title, global.StatusInProgress, order.ID)
	sendSuccess(c, gin.H{"message": "Work order taken successfully"})
}

// AssignStaff allows admin to assign a work order to a staff member
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

	assignee, err := repo.GetUserByID(input.AssigneeID)
	if err != nil {
		sendError(c, http.StatusNotFound, "Staff member not found")
		return
	}

	order, err := repo.GetWorkOrderById(orderID)
	if err != nil {
		sendError(c, http.StatusNotFound, "Work order not found")
		return
	}

	if order.Status == global.StatusCompleted {
		sendError(c, http.StatusBadRequest, "Work order already completed")
		return
	}

	if err := repo.AssignWorkOrder(orderID, input.AssigneeID); err != nil {
		log.Printf("Error assigning work order %d to user %d: %v", orderID, input.AssigneeID, err)
		sendError(c, http.StatusInternalServerError, "Failed to assign staff")
		return
	}

	repo.LogActivity(admin.ID, admin.Name, fmt.Sprintf("assigned request to %s:", assignee.Name), order.Title, global.StatusInProgress, order.ID)
	sendSuccess(c, gin.H{"message": "Staff assigned successfully"})
}

// FinalizeOrder marks a work order as completed
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
		sendError(c, http.StatusNotFound, "Work order not found")
		return
	}

	if order.AssigneeID == nil {
		sendError(c, http.StatusBadRequest, "Work order has not been assigned yet")
		return
	}

	// Check permission: only assignee or admin can finalize
	role, _ := c.Get("role")
	if *order.AssigneeID != user.ID && role != global.RoleAdmin {
		sendError(c, http.StatusForbidden, "Access denied")
		return
	}

	if err := repo.FinalizeWorkOrder(orderID, input.Note, user.ID); err != nil {
		log.Printf("Error finalizing work order %d: %v", orderID, err)
		sendError(c, http.StatusInternalServerError, "Failed to finalize work order")
		return
	}

	repo.LogActivity(user.ID, user.Name, "completed request:", order.Title, global.StatusCompleted, order.ID)
	sendSuccess(c, gin.H{"message": "Work order finalized successfully"})
}
