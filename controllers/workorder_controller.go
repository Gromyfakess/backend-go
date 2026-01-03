package controllers

import (
	"fmt"
	"siro-backend/constants"
	"siro-backend/models"
	"siro-backend/repository"
	"siro-backend/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Helper untuk mencatat Log Aktivitas ke Database
func logActivity(userID uint, userName, action, details, status string, reqID uint) {
	// Jalankan di background (goroutine) agar client tidak menunggu insert log selesai
	go func() {
		newLog := models.ActivityLog{
			UserID:    userID,
			UserName:  userName,
			Action:    action,
			Details:   details,
			Status:    status,
			RequestID: reqID,
			Timestamp: time.Now(),
		}
		// Abaikan error untuk log aktivitas agar tidak crash process utama
		_ = repository.CreateActivityLog(&newLog)
	}()
}

func GetActivities(c *gin.Context) {
	// Ambil 5 aktivitas terakhir
	logs := repository.GetRecentActivities(5)
	c.JSON(200, logs)
}

func CreateWorkOrder(c *gin.Context) {
	var input models.WorkOrderRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	uid, _ := c.Get("userID")
	role, _ := c.Get("role")
	canCRUD, _ := c.Get("canCRUD")

	canCRUDBool, ok := canCRUD.(bool)
	if !ok {
		canCRUDBool = false
	}

	if role != constants.RoleAdmin && !canCRUDBool {
		c.JSON(403, gin.H{"error": "Permission denied"})
		return
	}

	requester, _ := repository.GetUserByID(uid.(uint))

	if input.Unit == "" {
		c.JSON(400, gin.H{"error": "Unit tujuan harus dipilih"})
		return
	}

	newOrder := models.WorkOrder{
		Title:         input.Title,
		Description:   input.Description,
		Priority:      input.Priority,
		RequesterID:   requester.ID,
		RequesterName: requester.Name,
		Unit:          input.Unit,
		PhotoURL:      input.PhotoURL,
		Status:        constants.StatusPending,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := repository.CreateWorkOrder(&newOrder); err != nil {
		c.JSON(500, gin.H{"error": "Failed to create"})
		return
	}

	logDesc := fmt.Sprintf("membuat request ke %s:", input.Unit)
	logActivity(requester.ID, requester.Name, logDesc, newOrder.Title, constants.StatusPending, newOrder.ID)

	fullOrder, _ := repository.GetWorkOrderById(newOrder.ID)
	c.JSON(201, fullOrder)
}

func UploadWorkOrderEvidence(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "No file uploaded"})
		return
	}

	uploadConfig := utils.DefaultImageConfig("workorder")
	relativePath, err := utils.SaveUploadedFile(file, uploadConfig)

	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	fullURL := utils.GetBaseURL() + relativePath
	c.JSON(200, gin.H{
		"message": "File uploaded successfully",
		"url":     fullURL,
	})
}

func GetWorkOrders(c *gin.Context) {
	orders := repository.GetAllWorkOrders()
	c.JSON(200, orders)
}

func TakeRequest(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID format"})
		return
	}

	uid, _ := c.Get("userID")
	userID := uid.(uint)
	user, _ := repository.GetUserByID(userID)

	order, err := repository.GetWorkOrderById(uint(id))
	if err != nil {
		c.JSON(404, gin.H{"error": "Not found"})
		return
	}

	if order.Status == constants.StatusCompleted {
		c.JSON(400, gin.H{"error": "Tiket sudah selesai, tidak bisa diambil lagi."})
		return
	}

	now := time.Now()

	order.AssigneeID = &userID
	order.Status = constants.StatusInProgress
	order.TakenAt = &now

	repository.UpdateWorkOrder(&order)

	logActivity(user.ID, user.Name, "sedang mengerjakan:", order.Title, constants.StatusInProgress, order.ID)

	c.JSON(200, gin.H{"message": "Taken"})
}

func AssignStaff(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID format"})
		return
	}

	uid, _ := c.Get("userID")
	admin, _ := repository.GetUserByID(uid.(uint))

	var i models.AssignRequest
	if err := c.ShouldBindJSON(&i); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	assignee, err := repository.GetUserByID(i.AssigneeID)
	if err != nil {
		c.JSON(404, gin.H{"error": "Staff not found"})
		return
	}

	order, err := repository.GetWorkOrderById(uint(id))
	if err != nil {
		c.JSON(404, gin.H{"error": "Work order not found"})
		return
	}

	if order.Status == constants.StatusCompleted {
		c.JSON(400, gin.H{"error": "Tiket sudah selesai."})
		return
	}

	now := time.Now()

	order.AssigneeID = &i.AssigneeID
	order.Status = constants.StatusInProgress
	order.TakenAt = &now

	repository.UpdateWorkOrder(&order)

	logMessage := fmt.Sprintf("menugaskan request kepada %s:", assignee.Name)
	logActivity(admin.ID, admin.Name, logMessage, order.Title, constants.StatusInProgress, order.ID)

	c.JSON(200, gin.H{"message": "Assigned"})
}

func FinalizeOrder(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid ID format"})
		return
	}

	uid, _ := c.Get("userID")
	role, _ := c.Get("role")
	user, _ := repository.GetUserByID(uid.(uint))

	var input models.FinalizeRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		// Lanjut aja kalau bind gagal, note bisa kosong
	}

	order, err := repository.GetWorkOrderById(uint(id))
	if err != nil {
		c.JSON(404, gin.H{"error": "Not found"})
		return
	}

	if order.AssigneeID == nil {
		c.JSON(400, gin.H{"error": "Tiket belum diambil oleh siapapun"})
		return
	}

	if *order.AssigneeID != user.ID && role != constants.RoleAdmin {
		c.JSON(403, gin.H{"error": "Akses ditolak. Hanya staff penanggung jawab yang boleh menyelesaikan tiket."})
		return
	}

	now := time.Now()

	order.Status = constants.StatusCompleted
	order.CompletedAt = &now
	order.CompletedByID = &user.ID
	order.CompletionNote = input.Note

	repository.UpdateWorkOrder(&order)

	logActivity(user.ID, user.Name, "telah menyelesaikan tiket:", order.Title, constants.StatusCompleted, order.ID)

	c.JSON(200, gin.H{"message": "Done"})
}
