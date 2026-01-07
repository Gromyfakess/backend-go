package controllers

import (
	"fmt"
	"net/http"
	"siro-backend/constants" // Import constants
	"siro-backend/models"
	"siro-backend/repository"
	"siro-backend/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// ... (Kode GetActivities dan CreateWorkOrder sama, biarkan saja) ...
// Hanya tampilkan fungsi yang berubah untuk efisiensi, tapi pastikan package import di atas benar.

func GetActivities(c *gin.Context) {
	logs, err := repository.GetActivities()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch activities"})
		return
	}
	c.JSON(http.StatusOK, logs)
}

func CreateWorkOrder(c *gin.Context) {
	var input models.WorkOrderRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	uid, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized context"})
		return
	}

	role, _ := c.Get("role")
	canCRUD, _ := c.Get("canCRUD")
	canCRUDBool, _ := canCRUD.(bool)

	if role != constants.RoleAdmin && !canCRUDBool {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	if input.Unit == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unit tujuan harus dipilih"})
		return
	}

	newOrder := models.WorkOrder{
		Title:       input.Title,
		Description: input.Description,
		Priority:    input.Priority,
		RequesterID: uid.(uint),
		Unit:        input.Unit,
		PhotoURL:    input.PhotoURL,
		Status:      constants.StatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := repository.CreateWorkOrder(&newOrder); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create work order"})
		return
	}

	fullOrder, err := repository.GetWorkOrderById(newOrder.ID)
	if err == nil {
		logDesc := fmt.Sprintf("membuat request ke %s:", input.Unit)
		repository.LogActivity(fullOrder.RequesterID, fullOrder.RequesterName, logDesc, fullOrder.Title, constants.StatusPending, fullOrder.ID)
	}

	c.JSON(http.StatusCreated, fullOrder)
}

func UploadWorkOrderEvidence(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// FIX: Gunakan Constants
	uploadConfig := utils.DefaultImageConfig(constants.DirWorkOrder)

	relativePath, err := utils.SaveUploadedFile(file, uploadConfig)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fullURL := utils.GetBaseURL() + relativePath
	c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully", "url": fullURL})
}

// ... (Sisa fungsi GetWorkOrders, TakeRequest, AssignStaff, FinalizeOrder sama seperti sebelumnya) ...
func GetWorkOrders(c *gin.Context) {
	filters := map[string]string{
		"status":         c.Query("status"),
		"unit":           c.Query("unit"),
		"requester_unit": c.Query("requester_unit"),
		"date":           c.Query("date"),
	}

	orders, err := repository.GetWorkOrders(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch work orders"})
		return
	}
	if orders == nil {
		orders = []models.WorkOrder{}
	}
	c.JSON(http.StatusOK, orders)
}

func TakeRequest(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	uid, _ := c.Get("userID")
	userID := uid.(uint)

	user, err := repository.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	order, err := repository.GetWorkOrderById(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Work Order not found"})
		return
	}
	if order.Status == constants.StatusCompleted {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tiket sudah selesai"})
		return
	}

	if err := repository.TakeWorkOrder(uint(id), userID); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Gagal mengambil tiket. Mungkin sudah diambil orang lain."})
		return
	}

	repository.LogActivity(user.ID, user.Name, "sedang mengerjakan:", order.Title, constants.StatusInProgress, order.ID)
	c.JSON(http.StatusOK, gin.H{"message": "Taken successfully"})
}

func AssignStaff(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	uid, _ := c.Get("userID")
	admin, _ := repository.GetUserByID(uid.(uint))

	var i models.AssignRequest
	if err := c.ShouldBindJSON(&i); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	assignee, err := repository.GetUserByID(i.AssigneeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Staff not found"})
		return
	}

	order, err := repository.GetWorkOrderById(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Work order not found"})
		return
	}
	if order.Status == constants.StatusCompleted {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tiket sudah selesai."})
		return
	}

	if err := repository.AssignWorkOrder(uint(id), i.AssigneeID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign staff"})
		return
	}

	logMessage := fmt.Sprintf("menugaskan request kepada %s:", assignee.Name)
	repository.LogActivity(admin.ID, admin.Name, logMessage, order.Title, constants.StatusInProgress, order.ID)
	c.JSON(http.StatusOK, gin.H{"message": "Assigned successfully"})
}

func FinalizeOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	uid, _ := c.Get("userID")
	role, _ := c.Get("role")
	user, _ := repository.GetUserByID(uid.(uint))

	var input models.FinalizeRequest
	_ = c.ShouldBindJSON(&input)

	order, err := repository.GetWorkOrderById(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}

	if order.AssigneeID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tiket belum diambil"})
		return
	}

	if *order.AssigneeID != user.ID && role != constants.RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses ditolak"})
		return
	}

	if err := repository.FinalizeWorkOrder(uint(id), input.Note, user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to finalize order"})
		return
	}

	repository.LogActivity(user.ID, user.Name, "telah menyelesaikan tiket:", order.Title, constants.StatusCompleted, order.ID)
	c.JSON(http.StatusOK, gin.H{"message": "Work Order Finalized"})
}
