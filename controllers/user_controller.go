package controllers

import (
	"siro-backend/models"
	"siro-backend/repository"
	"siro-backend/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetMe(c *gin.Context) {
	uid, _ := c.Get("userID")
	user, err := repository.GetUserByID(uid.(uint))
	if err != nil {
		c.JSON(404, gin.H{"error": "User not found"})
		return
	}
	c.JSON(200, user)
}

func GetStaffList(c *gin.Context) {
	users := repository.GetStaffByUnit("IT Center")
	c.JSON(200, users)
}

func UpdateAvailability(c *gin.Context) {
	id := c.Param("id")
	var i models.AvailabilityRequest
	if err := c.ShouldBindJSON(&i); err != nil {
		c.JSON(400, gin.H{"error": "Invalid"})
		return
	}
	repository.UpdateAvailability(id, i.Status)
	c.JSON(200, gin.H{"status": "ok"})
}

func GetAllUsers(c *gin.Context) {
	users := repository.GetAllUsers()
	c.JSON(200, users)
}

func CreateUser(c *gin.Context) {
	var i models.UserRequest
	if err := c.ShouldBindJSON(&i); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	p, _ := utils.HashPassword(i.Password)
	u := models.User{
		Name:         i.Name,
		Email:        i.Email,
		Role:         i.Role,
		Unit:         i.Unit,
		CanCRUD:      i.CanCRUD,
		PasswordHash: p,
		Availability: "Offline",
		AvatarURL:    "https://i.pravatar.cc/150",
	}
	repository.CreateUser(&u)
	c.JSON(201, u)
}

func UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	var i models.UserRequest
	if err := c.ShouldBindJSON(&i); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	id, _ := strconv.Atoi(idStr)
	user, _ := repository.GetUserByID(uint(id))

	user.Name = i.Name
	user.Email = i.Email
	user.Role = i.Role
	user.Unit = i.Unit
	user.CanCRUD = i.CanCRUD
	if i.Password != "" {
		p, _ := utils.HashPassword(i.Password)
		user.PasswordHash = p
	}

	repository.UpdateUser(&user)
	c.JSON(200, user)
}

func DeleteUser(c *gin.Context) {
	id := c.Param("id")
	repository.DeleteUser(id)
	c.JSON(200, gin.H{"message": "Deleted"})
}
