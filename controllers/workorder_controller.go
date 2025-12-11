package controllers

import (
	"siro-backend/models"
	"siro-backend/repository"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

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
	unit := input.Unit
	if role != "Admin" {
		unit = requester.Unit
	}

	newOrder := models.WorkOrder{
		Title:         input.Title,
		Description:   input.Description,
		Priority:      input.Priority,
		RequesterID:   requester.ID,
		RequesterName: requester.Name, // Snapshot Nama
		// RequesterAvatar: SUDAH DIHAPUS (Diambil via Relasi RequesterData)
		Unit:      unit,
		Status:    "Pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := repository.CreateWorkOrder(&newOrder); err != nil {
		c.JSON(500, gin.H{"error": "Failed to create"})
		return
	}

	// Opsional: Reload data agar frontend mendapat struct lengkap (termasuk RequesterData)
	// Kita konversi ID uint ke string dulu untuk query
	fullOrder, _ := repository.GetWorkOrderById(strconv.Itoa(int(newOrder.ID)))

	c.JSON(201, fullOrder)
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

	order, err := repository.GetWorkOrderById(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Not found"})
		return
	}

	// Validasi kepemilikan (hanya admin atau pemilik request yang bisa edit)
	if role != "Admin" && order.RequesterID != uid.(uint) {
		c.JSON(403, gin.H{"error": "Forbidden"})
		return
	}

	order.Title = input.Title
	order.Description = input.Description
	order.Priority = input.Priority
	if role == "Admin" {
		order.Unit = input.Unit
	}

	repository.UpdateWorkOrder(&order)
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

	order, err := repository.GetWorkOrderById(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Not found"})
		return
	}

	order.AssigneeID = &userID
	order.Status = "In Progress"
	repository.UpdateWorkOrder(&order)
	c.JSON(200, gin.H{"message": "Taken"})
}

func AssignStaff(c *gin.Context) {
	id := c.Param("id")
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
	c.JSON(200, gin.H{"message": "Assigned"})
}

func FinalizeOrder(c *gin.Context) {
	id := c.Param("id")
	order, err := repository.GetWorkOrderById(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Not found"})
		return
	}
	order.Status = "Completed"
	repository.UpdateWorkOrder(&order)
	c.JSON(200, gin.H{"message": "Done"})
}
