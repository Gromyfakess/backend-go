package main

import (
	"fmt"
	"log"
	"os"
	"siro-backend/config"
	"siro-backend/controllers"
	"siro-backend/middleware"
	"siro-backend/models"
	"siro-backend/repository"
	"siro-backend/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func seedData() {
	var count int64
	config.DB.Model(&models.User{}).Count(&count)
	if count > 0 {
		return
	}
	users := []models.User{
		{Name: "David Hendrawan Ciu", Email: "david@uib.ac.id", Role: "Admin", Unit: "IT Center", CanCRUD: true, Availability: "Online", AvatarURL: "https://i.pravatar.cc/150?u=99", PasswordHash: "$2a$12$eyHuyvSrUOPK2SLd/aZFd.CJWgZ0cUYySCTQQd8n5ce2qSbfJ9vJK"},
		{Name: "Aulia Maharani Hasan", Email: "aulia@uib.ac.id", Role: "Staff", Unit: "IT Center", CanCRUD: false, Availability: "Online", AvatarURL: "https://i.pravatar.cc/150?u=1", PasswordHash: "$2a$12$.7b5.b0Wc96YH4yGtqFus.kC4OzabdKm2y.WWAaDgUI.dumw4SUVm"},
	}
	repository.CreateUser(&users[0])
	repository.CreateUser(&users[1])
	fmt.Println("Seeding Data Done!")
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found")
	}

	// 2. Init Utils & Config
	utils.InitJWT()
	config.ConnectDB()
	seedData()

	r := gin.Default()

	// 3. Setup CORS dinamis
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{frontendURL},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	// ... (Routes Group TETAP SAMA seperti sebelumnya) ...
	// Copy-paste bagian r.POST, r.Group, api.Use, dsb dari main.go lama
	r.POST("/login", controllers.LoginHandler)
	r.POST("/refresh", controllers.RefreshHandler)
	r.POST("/logout", controllers.LogoutHandler)

	api := r.Group("/")
	api.Use(middleware.AuthMiddleware())
	{
		api.GET("/me", controllers.GetMe)
		api.GET("/staff", controllers.GetStaffList)
		api.PATCH("/staff/:id/availability", controllers.UpdateAvailability)

		api.GET("/workorders", controllers.GetWorkOrders)
		api.POST("/workorders", controllers.CreateWorkOrder)
		api.PUT("/workorders/:id", controllers.UpdateWorkOrder)
		api.DELETE("/workorders/:id", controllers.DeleteWorkOrder)

		api.PATCH("/workorders/:id/take", controllers.TakeRequest)
		api.PATCH("/workorders/:id/assign", controllers.AssignStaff)
		api.PATCH("/workorders/:id/finalize", controllers.FinalizeOrder)

		users := api.Group("/users")
		users.Use(middleware.AdminOnly())
		{
			users.GET("", controllers.GetAllUsers)
			users.POST("", controllers.CreateUser)
			users.PUT("/:id", controllers.UpdateUser)
			users.DELETE("/:id", controllers.DeleteUser)
		}
	}

	// 4. Run Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Server running on http://127.0.0.1:%s\n", port)
	r.Run(":" + port)
}
