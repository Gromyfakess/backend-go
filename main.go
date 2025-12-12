package main

import (
	"fmt"
	"log"
	"os"
	"siro-backend/config"
	"siro-backend/controllers"
	"siro-backend/middleware"
	"siro-backend/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found")
	}

	utils.InitJWT()
	config.ConnectDB()

	r := gin.Default()

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

	// 1. STATIC FILE SERVING (Agar gambar yang diupload bisa dibuka)
	// Akses via: http://localhost:8080/uploads/namafile.jpg
	r.Static("/uploads", "./uploads")

	// Public Routes
	r.POST("/login", controllers.LoginHandler)
	r.POST("/refresh", controllers.RefreshHandler)
	r.POST("/logout", controllers.LogoutHandler)

	// Protected Routes
	api := r.Group("/")
	api.Use(middleware.AuthMiddleware())
	{
		api.GET("/me", controllers.GetMe)
		api.PUT("/me", controllers.UpdateMe)
		api.POST("/upload", controllers.UploadFile)

		// Staff & Status
		api.GET("/staff", controllers.GetStaffList)
		api.PATCH("/staff/:id/availability", controllers.UpdateAvailability)

		// Work Order
		api.GET("/workorders", controllers.GetWorkOrders)
		api.POST("/workorders", controllers.CreateWorkOrder)
		api.PUT("/workorders/:id", controllers.UpdateWorkOrder)
		api.DELETE("/workorders/:id", controllers.DeleteWorkOrder)

		api.PATCH("/workorders/:id/take", controllers.TakeRequest)
		api.PATCH("/workorders/:id/assign", controllers.AssignStaff)
		api.PATCH("/workorders/:id/finalize", controllers.FinalizeOrder)

		// Admin User Management
		users := api.Group("/users")
		users.Use(middleware.AdminOnly())
		{
			users.GET("", controllers.GetAllUsers)
			users.POST("", controllers.CreateUser)
			users.PUT("/:id", controllers.UpdateUser)
			users.DELETE("/:id", controllers.DeleteUser)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Server running on http://127.0.0.1:%s\n", port)
	r.Run(":" + port)
}
