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

// Helper untuk log
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

// HANDLER BARU: Get Activities (Solusi Error 404)
func GetActivities(c *gin.Context) {
	logs := repository.GetRecentActivities()
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

	if role != "Admin" && canCRUD == false {
		c.JSON(403, gin.H{"error": "Permission denied"})
		return
	}

	requester, _ := repository.GetUserByID(uid.(uint))
	unit := input.Unit
	if role != "Admin" {
		unit = requester.Unit
	}

	newOrder := models.WorkOrder{
		Title:         input.Title,
		Description:   input.Description,
		Priority:      input.Priority,
		RequesterID:   requester.ID,
		RequesterName: requester.Name,
		Unit:          unit,
		PhotoURL:      input.PhotoURL, // Simpan Foto
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

	fullOrder, _ := repository.GetWorkOrderById(strconv.Itoa(int(newOrder.ID)))
	c.JSON(201, fullOrder)
}

func UploadWorkOrderEvidence(c *gin.Context) {
	// 1. Ambil file dari form-data
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "No file uploaded"})
		return
	}

	// 2. Tentukan folder tujuan: uploads/workorder
	// Pastikan folder ini ada, jika tidak buat foldernya
	uploadPath := "uploads/workorder"
	if _, err := os.Stat(uploadPath); os.IsNotExist(err) {
		os.MkdirAll(uploadPath, os.ModePerm)
	}

	// 3. Buat nama file unik (Timestamp + OriginalName) agar tidak bentrok
	// Gunakan filepath.Base untuk keamanan (menghindari path traversal)
	filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(file.Filename))
	dst := filepath.Join(uploadPath, filename)

	// 4. Simpan file
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(500, gin.H{"error": "Failed to save file"})
		return
	}

	// 5. Generate URL Publik
	// Karena di main.go folder "uploads" di-serve sebagai static,
	// URL-nya menjadi /uploads/workorder/namafile
	// (Perhatikan slash '/' manual agar kompatibel dengan URL browser)
	fileURL := "/uploads/workorder/" + filename

	c.JSON(200, gin.H{
		"message": "File uploaded successfully",
		"url":     fileURL,
	})
}

func GetWorkOrders(c *gin.Context) {
	orders := repository.GetAllWorkOrders()
	c.JSON(200, orders)
}

func UpdateWorkOrder(c *gin.Context) {
	id := c.Param("id")
	var input models.WorkOrderRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	uid, _ := c.Get("userID")
	role, _ := c.Get("role")

	// Fetch data User/Admin yang sedang melakukan aksi untuk dicatat namanya di log
	actorID := uid.(uint)
	actor, _ := repository.GetUserByID(actorID)

	order, err := repository.GetWorkOrderById(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Not found"})
		return
	}

	// Hanya Admin atau Pemilik Tiket yang boleh edit
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

	// Menentukan pesan aksi berdasarkan Role
	actionMsg := "memperbarui tiket:"
	if role == "Admin" {
		actionMsg = "[Admin] memperbarui data:"
	}

	logActivity(actor.ID, actor.Name, actionMsg, order.Title, order.Status)

	c.JSON(200, order)
}

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

	order.AssigneeID = &userID
	order.Status = "In Progress"
	repository.UpdateWorkOrder(&order)

	logActivity(user.ID, user.Name, "sedang mengerjakan:", order.Title, "In Progress")

	c.JSON(200, gin.H{"message": "Taken"})
}

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

	order.AssigneeID = &i.AssigneeID
	order.Status = "In Progress"
	repository.UpdateWorkOrder(&order)

	// LOG AKTIVITAS
	logActivity(admin.ID, admin.Name, "menugaskan tiket:", order.Title, "In Progress")

	c.JSON(200, gin.H{"message": "Assigned"})
}

func FinalizeOrder(c *gin.Context) {
	id := c.Param("id")
	uid, _ := c.Get("userID")
	user, _ := repository.GetUserByID(uid.(uint))

	order, err := repository.GetWorkOrderById(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Not found"})
		return
	}
	order.Status = "Completed"
	repository.UpdateWorkOrder(&order)

	// LOG AKTIVITAS
	logActivity(user.ID, user.Name, "telah menyelesaikan:", order.Title, "Completed")

	c.JSON(200, gin.H{"message": "Done"})
}
