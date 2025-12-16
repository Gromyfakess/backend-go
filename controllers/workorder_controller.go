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
func logActivity(userID uint, userName, action, details, status string) {
	newLog := models.ActivityLog{
		UserID:    userID,
		UserName:  userName,
		Action:    action,
		Details:   details,
		Status:    status,
		Timestamp: time.Now(),
	}
	repository.CreateActivityLog(&newLog)
}

// GetActivities: Mengambil log aktivitas terbaru untuk Dashboard
func GetActivities(c *gin.Context) {
	logs := repository.GetRecentActivities()
	c.JSON(200, logs)
}

// CreateWorkOrder: Membuat tiket baru
func CreateWorkOrder(c *gin.Context) {
	var input models.WorkOrderRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	uid, _ := c.Get("userID")
	role, _ := c.Get("role")
	canCRUD, _ := c.Get("canCRUD")

	// Validasi Permission: Hanya Admin atau Staff dengan 'canCRUD' yang boleh buat
	if role != "Admin" && canCRUD == false {
		c.JSON(403, gin.H{"error": "Permission denied"})
		return
	}

	requester, _ := repository.GetUserByID(uid.(uint))
	unit := input.Unit
	if role != "Admin" {
		unit = requester.Unit // User biasa hanya bisa buat tiket untuk unitnya sendiri
	}

	newOrder := models.WorkOrder{
		Title:         input.Title,
		Description:   input.Description,
		Priority:      input.Priority,
		RequesterID:   requester.ID,
		RequesterName: requester.Name,
		Unit:          unit,
		PhotoURL:      input.PhotoURL,
		Status:        "Pending",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := repository.CreateWorkOrder(&newOrder); err != nil {
		c.JSON(500, gin.H{"error": "Failed to create"})
		return
	}

	// LOG AKTIVITAS
	logActivity(requester.ID, requester.Name, "telah membuat request:", newOrder.Title, "Pending")

	// Return full object response
	fullOrder, _ := repository.GetWorkOrderById(strconv.Itoa(int(newOrder.ID)))
	c.JSON(201, fullOrder)
}

// UploadWorkOrderEvidence: Upload foto bukti tiket
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

// GetWorkOrders: Mengambil semua tiket
func GetWorkOrders(c *gin.Context) {
	orders := repository.GetAllWorkOrders()
	c.JSON(200, orders)
}

// UpdateWorkOrder: Edit tiket (Judul, Deskripsi, dll)
func UpdateWorkOrder(c *gin.Context) {
	id := c.Param("id")
	var input models.WorkOrderRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	uid, _ := c.Get("userID")
	role, _ := c.Get("role")

	actorID := uid.(uint)
	actor, _ := repository.GetUserByID(actorID)

	order, err := repository.GetWorkOrderById(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Not found"})
		return
	}

	// Permission: Hanya Admin atau Pemilik Tiket yang boleh edit
	if role != "Admin" && order.RequesterID != actorID {
		c.JSON(403, gin.H{"error": "Forbidden"})
		return
	}

	order.Title = input.Title
	order.Description = input.Description
	order.Priority = input.Priority

	if input.PhotoURL != "" {
		order.PhotoURL = input.PhotoURL
	}

	if role == "Admin" {
		order.Unit = input.Unit
	}

	repository.UpdateWorkOrder(&order)

	// LOG AKTIVITAS (Update)
	actionMsg := "memperbarui tiket:"
	if role == "Admin" {
		actionMsg = "Admin memperbarui data:"
	}
	logActivity(actor.ID, actor.Name, actionMsg, order.Title, order.Status)

	c.JSON(200, order)
}

// DeleteWorkOrder: Hapus tiket
func DeleteWorkOrder(c *gin.Context) {
	id := c.Param("id")
	uid, _ := c.Get("userID")
	role, _ := c.Get("role")

	order, err := repository.GetWorkOrderById(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Not found"})
		return
	}

	if role != "Admin" && order.RequesterID != uid.(uint) {
		c.JSON(403, gin.H{"error": "Forbidden"})
		return
	}

	repository.DeleteWorkOrder(&order)
	c.JSON(200, gin.H{"message": "Deleted"})
}

// TakeRequest: Staff mengambil tiket sendiri
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
	logActivity(user.ID, user.Name, "sedang mengerjakan:", order.Title, "In Progress")

	c.JSON(200, gin.H{"message": "Taken"})
}

// AssignStaff: Admin menugaskan tiket ke staff lain
func AssignStaff(c *gin.Context) {
	id := c.Param("id")
	uid, _ := c.Get("userID")
	admin, _ := repository.GetUserByID(uid.(uint))

	var i models.AssignRequest
	if err := c.ShouldBindJSON(&i); err != nil {
		c.JSON(400, gin.H{"error": "Invalid"})
		return
	}

	order, err := repository.GetWorkOrderById(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Not found"})
		return
	}

	now := time.Now()

	order.AssigneeID = &i.AssigneeID
	order.Status = "In Progress"
	order.TakenAt = &now // Simpan waktu TakenAt

	repository.UpdateWorkOrder(&order)

	// LOG AKTIVITAS
	logActivity(admin.ID, admin.Name, "menugaskan tiket:", order.Title, "In Progress")

	c.JSON(200, gin.H{"message": "Assigned"})
}

// FinalizeOrder: Menyelesaikan tiket dengan catatan
func FinalizeOrder(c *gin.Context) {
	id := c.Param("id")
	uid, _ := c.Get("userID")
	user, _ := repository.GetUserByID(uid.(uint))

	// Ambil Note dari body request (models.FinalizeRequest harus ada di models.go)
	var input models.FinalizeRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		// Note opsional, lanjut saja jika kosong/error binding
	}

	order, err := repository.GetWorkOrderById(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Not found"})
		return
	}

	now := time.Now()

	order.Status = "Completed"
	order.CompletedAt = &now
	order.CompletedByID = &user.ID
	order.CompletionNote = input.Note

	repository.UpdateWorkOrder(&order)

	logActivity(user.ID, user.Name, "telah menyelesaikan tiket:", order.Title, "Completed")

	c.JSON(200, gin.H{"message": "Done"})
}
