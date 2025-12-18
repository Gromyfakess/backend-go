package controllers

import (
	"fmt"
	"os"
	"path/filepath"
	"siro-backend/models"
	"siro-backend/repository"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Helper untuk mencatat Log Aktivitas ke Database
func logActivity(userID uint, userName, action, details, status string, reqID uint) {
	newLog := models.ActivityLog{
		UserID:    userID,
		UserName:  userName,
		Action:    action,
		Details:   details,
		Status:    status,
		RequestID: reqID,
		Timestamp: time.Now(),
	}
	repository.CreateActivityLog(&newLog)
}

// GetActivities: Mengambil log aktivitas terbaru untuk Dashboard
func GetActivities(c *gin.Context) {
	logs := repository.GetRecentActivities()
	c.JSON(200, logs)
}

// CreateWorkOrder: Membuat request baru
func CreateWorkOrder(c *gin.Context) {
	var input models.WorkOrderRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	uid, _ := c.Get("userID")
	role, _ := c.Get("role")
	canCRUD, _ := c.Get("canCRUD")

	// Validasi Permission
	if role != "Admin" && canCRUD == false {
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
		Unit:          input.Unit, // Menggunakan Unit Tujuan dari Frontend
		PhotoURL:      input.PhotoURL,
		Status:        "Pending",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := repository.CreateWorkOrder(&newOrder); err != nil {
		c.JSON(500, gin.H{"error": "Failed to create"})
		return
	}

	logDesc := fmt.Sprintf("membuat request ke %s:", input.Unit)
	logActivity(requester.ID, requester.Name, logDesc, newOrder.Title, "Pending", newOrder.ID)

	fullOrder, _ := repository.GetWorkOrderById(strconv.Itoa(int(newOrder.ID)))
	c.JSON(201, fullOrder)
}

// UploadWorkOrderEvidence: Upload foto bukti untuk Request
func UploadWorkOrderEvidence(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "No file uploaded"})
		return
	}

	uploadPath := "uploads/workorder"
	if _, err := os.Stat(uploadPath); os.IsNotExist(err) {
		os.MkdirAll(uploadPath, os.ModePerm)
	}

	// Nama file unik
	filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(file.Filename))
	dst := filepath.Join(uploadPath, filename)

	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(500, gin.H{"error": "Failed to save file"})
		return
	}

	fileURL := "/uploads/workorder/" + filename
	c.JSON(200, gin.H{
		"message": "File uploaded successfully",
		"url":     fileURL,
	})
}

// GetWorkOrders: Mengambil semua request
func GetWorkOrders(c *gin.Context) {
	orders := repository.GetAllWorkOrders()
	c.JSON(200, orders)
}

// TakeRequest: Staff mengambil request sendiri
func TakeRequest(c *gin.Context) {
	id := c.Param("id")
	uid, _ := c.Get("userID")
	userID := uid.(uint)
	user, _ := repository.GetUserByID(userID)

	order, err := repository.GetWorkOrderById(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Not found"})
		return
	}

	now := time.Now() // Waktu mulai pengerjaan

	order.AssigneeID = &userID
	order.Status = "In Progress"
	order.TakenAt = &now // Simpan waktu TakenAt

	repository.UpdateWorkOrder(&order)

	// LOG AKTIVITAS
	logActivity(user.ID, user.Name, "sedang mengerjakan:", order.Title, "In Progress", order.ID)

	c.JSON(200, gin.H{"message": "Taken"})
}

// AssignStaff: Admin menugaskan request ke staff lain
func AssignStaff(c *gin.Context) {
	id := c.Param("id")
	uid, _ := c.Get("userID")
	admin, _ := repository.GetUserByID(uid.(uint))

	var i models.AssignRequest
	if err := c.ShouldBindJSON(&i); err != nil {
		c.JSON(400, gin.H{"error": "Invalid"})
		return
	}

	assignee, err := repository.GetUserByID(i.AssigneeID)
	if err != nil {
		c.JSON(404, gin.H{"error": "Staff not found"})
		return
	}

	order, err := repository.GetWorkOrderById(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Work order not found"})
		return
	}

	now := time.Now()

	order.AssigneeID = &i.AssigneeID
	order.Status = "In Progress"
	order.TakenAt = &now

	repository.UpdateWorkOrder(&order)

	logMessage := fmt.Sprintf("menugaskan request kepada %s:", assignee.Name)
	logActivity(admin.ID, admin.Name, logMessage, order.Title, "In Progress", order.ID)

	c.JSON(200, gin.H{"message": "Assigned"})
}

// FinalizeOrder: Menyelesaikan tiket dengan catatan
// FinalizeOrder: Menyelesaikan tiket dengan catatan
func FinalizeOrder(c *gin.Context) {
	id := c.Param("id")
	uid, _ := c.Get("userID")
	role, _ := c.Get("role")
	user, _ := repository.GetUserByID(uid.(uint))

	var input models.FinalizeRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		// Note opsional
	}

	order, err := repository.GetWorkOrderById(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Not found"})
		return
	}

	// 1. Cek apakah tiket sudah diambil?
	if order.AssigneeID == nil {
		c.JSON(400, gin.H{"error": "Tiket belum diambil (Take) oleh siapapun"})
		return
	}

	// 2. Validasi: Hanya Assignee atau Admin yang boleh finalize
	if *order.AssigneeID != user.ID && role != "Admin" {
		c.JSON(403, gin.H{"error": "Anda tidak memiliki akses. Hanya staff yang mengerjakan tiket ini yang dapat menyelesaikannya."})
		return
	}

	now := time.Now()

	order.Status = "Completed"
	order.CompletedAt = &now
	order.CompletedByID = &user.ID
	order.CompletionNote = input.Note

	repository.UpdateWorkOrder(&order)

	logActivity(user.ID, user.Name, "telah menyelesaikan tiket:", order.Title, "Completed", order.ID)

	c.JSON(200, gin.H{"message": "Done"})
}
