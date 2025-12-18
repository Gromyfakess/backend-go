package main

import (
	"fmt"
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
	// Load .env diabaikan jika error (untuk production)
	_ = godotenv.Load()

	utils.InitJWT()

	// Ini sekarang tidak akan memblokir startup lama-lama
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

	// Static files (Untuk production disarankan pakai Cloudinary, tapi ini ok untuk fallback)
	r.Static("/uploads", "./uploads")

	// --- PENTING: Health Check Route ---
	// Route ini ringan, hanya untuk memberi tahu Leapcell bahwa app sudah hidup
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "alive",
			"message": "Siro Backend is Running",
		})
	})

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

		api.GET("/staff", controllers.GetStaffList)
		api.PATCH("/staff/:id/availability", controllers.UpdateAvailability)

		api.GET("/workorders", controllers.GetWorkOrders)
		api.POST("/workorders", controllers.CreateWorkOrder)
		api.POST("/upload/workorder", controllers.UploadWorkOrderEvidence)

		api.PATCH("/workorders/:id/take", controllers.TakeRequest)
		api.PATCH("/workorders/:id/assign", controllers.AssignStaff)
		api.PATCH("/workorders/:id/finalize", controllers.FinalizeOrder)

		api.GET("/activities", controllers.GetActivities)

		// Admin
		admin := api.Group("/admin")
		admin.Use(middleware.AdminOnly())
		{
			admin.GET("/users", controllers.GetAllUsers)
			admin.POST("/users", controllers.CreateUser)
			admin.PUT("/users/:id", controllers.UpdateUser)
			admin.DELETE("/users/:id", controllers.DeleteUser)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Server running on port :%s\n", port)

	// Pastikan listen di 0.0.0.0 (Gin defaultnya sudah ini jika pakai ":port")
	r.Run(":" + port)
}
