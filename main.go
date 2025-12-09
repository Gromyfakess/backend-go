package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var jwtSecret = []byte("SUPER_SECRET_KEY_PRODUCTION_SIRA")

// --- MODELS ---

type User struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name"`
	Email        string    `json:"email" gorm:"unique;size:100"`
	Role         string    `json:"role"`
	Unit         string    `json:"unit"`
	AvatarURL    string    `json:"avatar"`
	Availability string    `json:"availability"`
	CanCRUD      bool      `json:"canCRUD"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"-"`
}

type RefreshToken struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    uint   `gorm:"index"`
	Token     string `gorm:"type:varchar(500);index"`
	ExpiresAt time.Time
}

type WorkOrder struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
	Status      string `json:"status"`

	// Snapshot Data Requester (Agar jika user dihapus/ganti nama, history tetap ada)
	RequesterID     uint   `json:"requesterId"`
	Requester       string `json:"requester"`
	RequesterAvatar string `json:"requesterAvatar"` // <-- FIELD BARU (Menyimpan URL Foto)
	Unit            string `json:"unit"`

	AssigneeID *uint `json:"assigneeId"`
	Assignee   User  `json:"assignee" gorm:"foreignKey:AssigneeID"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// --- REQUEST STRUCTS ---

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UserRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password"`
	Role     string `json:"role" binding:"required"`
	Unit     string `json:"unit" binding:"required"`
	CanCRUD  bool   `json:"canCRUD"`
}

type WorkOrderRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Priority    string `json:"priority" binding:"required"`
	Unit        string `json:"unit"`
	AssigneeID  uint   `json:"assigneeId"`
}

type AssignRequest struct {
	AssigneeID uint `json:"assigneeId" binding:"required"`
}

type AvailabilityRequest struct {
	Status string `json:"status" binding:"required"`
}

var db *gorm.DB

// --- DATABASE CONNECTION ---

func connectDB() {
	// GANTI user:password SESUAI CONFIG MYSQL ANDA
	dsn := "root:@tcp(127.0.0.1:3306)/workorder_db?charset=utf8mb4&parseTime=True&loc=Local"
	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("DB Error:", err)
	}
	// Auto Migrate (Akan menambah kolom requester_avatar otomatis)
	db.AutoMigrate(&User{}, &RefreshToken{}, &WorkOrder{})
}

func seedData() {
	var count int64
	db.Model(&User{}).Count(&count)
	if count > 0 {
		return
	}

	log.Println("Seeding Data...")
	pass, _ := hashPassword("admin123")

	users := []User{
		{Name: "Super Admin", Email: "admin@uib.ac.id", Role: "Admin", Unit: "IT Center", CanCRUD: true, Availability: "Online", AvatarURL: "https://i.pravatar.cc/150?u=99", PasswordHash: pass},
		{Name: "Mike Chen", Email: "mike@uib.ac.id", Role: "Staff", Unit: "IT Center", CanCRUD: false, Availability: "Online", AvatarURL: "https://i.pravatar.cc/150?u=1", PasswordHash: pass},
		{Name: "Sarah Wilson", Email: "sarah@uib.ac.id", Role: "Staff", Unit: "IT Center", CanCRUD: true, Availability: "Busy", AvatarURL: "https://i.pravatar.cc/150?u=2", PasswordHash: pass},
		{Name: "David Park", Email: "david@uib.ac.id", Role: "Staff", Unit: "IT Center", CanCRUD: false, Availability: "Away", AvatarURL: "https://i.pravatar.cc/150?u=3", PasswordHash: pass},
	}
	db.Create(&users)
}

// --- UTILS ---

func hashPassword(p string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(p), 14)
	return string(b), err
}

func generateTokens(user User) (string, string, error) {
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID, "role": user.Role, "canCRUD": user.CanCRUD, "exp": time.Now().Add(time.Minute * 60).Unix(),
	})
	accessToken, _ := at.SignedString(jwtSecret)

	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID, "exp": time.Now().Add(time.Hour * 24 * 7).Unix(),
	})
	refreshToken, _ := rt.SignedString(jwtSecret)
	return accessToken, refreshToken, nil
}

// --- MIDDLEWARE ---

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}
		tokenString := strings.Split(authHeader, "Bearer ")[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) { return jwtSecret, nil })

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(401, gin.H{"error": "Invalid Token"})
			return
		}
		claims, _ := token.Claims.(jwt.MapClaims)
		c.Set("userID", uint(claims["user_id"].(float64)))
		c.Set("role", claims["role"].(string))
		c.Set("canCRUD", claims["canCRUD"].(bool))
		c.Next()
	}
}

func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		if role, _ := c.Get("role"); role != "Admin" {
			c.AbortWithStatusJSON(403, gin.H{"error": "Admin only"})
			return
		}
		c.Next()
	}
}

// --- AUTH HANDLERS ---

func loginHandler(c *gin.Context) {
	var input LoginRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}
	var user User
	if err := db.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(401, gin.H{"error": "User not found"})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)) != nil {
		c.JSON(401, gin.H{"error": "Wrong password"})
		return
	}

	acc, ref, _ := generateTokens(user)
	db.Create(&RefreshToken{UserID: user.ID, Token: ref, ExpiresAt: time.Now().Add(time.Hour * 24 * 7)})
	c.SetCookie("refresh_token", ref, 3600*24*7, "/", "127.0.0.1", false, true)
	c.JSON(200, gin.H{"accessToken": acc, "user": user})
}

func refresh(c *gin.Context) {
	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(401, gin.H{"error": "No token"})
		return
	}
	token, err := jwt.Parse(cookie, func(t *jwt.Token) (interface{}, error) { return jwtSecret, nil })
	if err != nil || !token.Valid {
		c.JSON(401, gin.H{"error": "Invalid"})
		return
	}

	claims := token.Claims.(jwt.MapClaims)
	var u User
	if err := db.First(&u, uint(claims["user_id"].(float64))).Error; err != nil {
		c.JSON(401, gin.H{"error": "User invalid"})
		return
	}

	acc, _, _ := generateTokens(u)
	c.JSON(200, gin.H{"accessToken": acc})
}

func logout(c *gin.Context) {
	c.SetCookie("refresh_token", "", -1, "/", "127.0.0.1", false, true)
	c.JSON(200, gin.H{"msg": "Bye"})
}

// --- WORK ORDER HANDLERS ---

func createWorkOrder(c *gin.Context) {
	var input WorkOrderRequest
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

	var requester User
	db.First(&requester, uid)

	unit := input.Unit
	if role != "Admin" {
		unit = requester.Unit
	}

	newOrder := WorkOrder{
		Title: input.Title, Description: input.Description, Priority: input.Priority,
		RequesterID:     requester.ID,
		Requester:       requester.Name,
		RequesterAvatar: requester.AvatarURL, // <--- MENYIMPAN FOTO
		Unit:            unit, Status: "Pending",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}

	if input.AssigneeID != 0 {
		newOrder.AssigneeID = &input.AssigneeID
		newOrder.Status = "In Progress"
	}

	db.Create(&newOrder)
	// Return result dengan data assignee ter-load
	db.Preload("Assignee").First(&newOrder, newOrder.ID)
	c.JSON(201, newOrder)
}

func updateWorkOrder(c *gin.Context) {
	id := c.Param("id")
	uid, _ := c.Get("userID")
	role, _ := c.Get("role")

	var order WorkOrder
	if err := db.First(&order, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Not found"})
		return
	}

	if role != "Admin" && order.RequesterID != uid.(uint) {
		c.JSON(403, gin.H{"error": "Forbidden"})
		return
	}

	var input WorkOrderRequest
	c.ShouldBindJSON(&input)

	order.Title = input.Title
	order.Description = input.Description
	order.Priority = input.Priority
	if role == "Admin" {
		order.Unit = input.Unit
	}

	db.Save(&order)
	c.JSON(200, order)
}

func deleteWorkOrder(c *gin.Context) {
	id := c.Param("id")
	uid, _ := c.Get("userID")
	role, _ := c.Get("role")

	var order WorkOrder
	if err := db.First(&order, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Not found"})
		return
	}

	if role != "Admin" && order.RequesterID != uid.(uint) {
		c.JSON(403, gin.H{"error": "Forbidden"})
		return
	}

	db.Delete(&order)
	c.JSON(200, gin.H{"message": "Deleted"})
}

func takeRequest(c *gin.Context) {
	id := c.Param("id")
	uid, _ := c.Get("userID")
	userID := uid.(uint)

	var order WorkOrder
	if err := db.First(&order, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Not found"})
		return
	}

	order.AssigneeID = &userID
	order.Status = "In Progress"
	db.Save(&order)
	c.JSON(200, gin.H{"message": "Taken"})
}

func sendStaff(c *gin.Context) {
	id := c.Param("id")
	var i AssignRequest
	if err := c.ShouldBindJSON(&i); err != nil {
		c.JSON(400, gin.H{"error": "Invalid"})
		return
	}

	var order WorkOrder
	if err := db.First(&order, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Not found"})
		return
	}

	order.AssigneeID = &i.AssigneeID
	order.Status = "In Progress"
	db.Save(&order)
	c.JSON(200, gin.H{"message": "Assigned"})
}

func finalizeOrder(c *gin.Context) {
	db.Model(&WorkOrder{}).Where("id = ?", c.Param("id")).Update("status", "Completed")
	c.JSON(200, gin.H{"message": "Done"})
}

func getWorkOrders(c *gin.Context) {
	var o []WorkOrder
	db.Preload("Assignee").Order("created_at desc").Find(&o)
	c.JSON(200, o)
}

// --- USER & GENERAL HANDLERS ---

func getStaffList(c *gin.Context) {
	var u []User
	// Hanya return role Staff atau semua? Untuk dropdown send staff sebaiknya return semua staff IT
	db.Where("unit = ?", "IT Center").Find(&u)
	c.JSON(200, u)
}

func me(c *gin.Context) {
	uid, _ := c.Get("userID")
	var u User
	db.First(&u, uid)
	c.JSON(200, u)
}

func updateAvail(c *gin.Context) {
	id := c.Param("id")
	var i AvailabilityRequest
	c.ShouldBindJSON(&i)
	db.Model(&User{}).Where("id = ?", id).Update("availability", i.Status)
	c.JSON(200, gin.H{"status": "ok"})
}

// --- ADMIN USER MANAGEMENT HANDLERS ---

func getAllUsers(c *gin.Context) { var u []User; db.Find(&u); c.JSON(200, u) }

func createUser(c *gin.Context) {
	var i UserRequest
	if err := c.ShouldBindJSON(&i); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	p, _ := hashPassword(i.Password)
	u := User{Name: i.Name, Email: i.Email, Role: i.Role, Unit: i.Unit, CanCRUD: i.CanCRUD, PasswordHash: p, Availability: "Offline", AvatarURL: "https://i.pravatar.cc/150"}
	if err := db.Create(&u).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed"})
		return
	}
	c.JSON(201, u)
}

func updateUser(c *gin.Context) {
	id := c.Param("id")
	var i UserRequest
	if err := c.ShouldBindJSON(&i); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	var u User
	if err := db.First(&u, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Not found"})
		return
	}

	u.Name = i.Name
	u.Email = i.Email
	u.Role = i.Role
	u.Unit = i.Unit
	u.CanCRUD = i.CanCRUD
	if i.Password != "" {
		p, _ := hashPassword(i.Password)
		u.PasswordHash = p
	}
	db.Save(&u)
	c.JSON(200, u)
}

func deleteUser(c *gin.Context) {
	db.Delete(&User{}, c.Param("id"))
	c.JSON(200, gin.H{"message": "Deleted"})
}

// --- MAIN ---

func main() {
	connectDB()
	seedData()

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	r.POST("/login", loginHandler)
	r.POST("/refresh", refresh)
	r.POST("/logout", logout)

	api := r.Group("/")
	api.Use(AuthMiddleware())
	{
		api.GET("/me", me)
		api.GET("/staff", getStaffList) // Buat dropdown staff
		api.PATCH("/staff/:id/availability", updateAvail)

		// Work Order
		api.GET("/workorders", getWorkOrders)
		api.POST("/workorders", createWorkOrder)
		api.PUT("/workorders/:id", updateWorkOrder)
		api.DELETE("/workorders/:id", deleteWorkOrder)

		api.PATCH("/workorders/:id/take", takeRequest)
		api.PATCH("/workorders/:id/assign", sendStaff)
		api.PATCH("/workorders/:id/finalize", finalizeOrder)

		// User Management
		users := api.Group("/users")
		users.Use(AdminOnly())
		{
			users.GET("", getAllUsers)
			users.POST("", createUser)
			users.PUT("/:id", updateUser)
			users.DELETE("/:id", deleteUser)
		}
	}
	fmt.Println("Server running on http://127.0.0.1:8080")
	r.Run("127.0.0.1:8080")
}
