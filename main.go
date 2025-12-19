package main

import (
	"os"
	"siro-backend/config"
	"siro-backend/controllers"
	"siro-backend/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Load .env file (Hanya untuk local development)
	// Di Leapcell, environment variable diset lewat dashboard, jadi error diabaikan
	_ = godotenv.Load()

	// 2. Connect ke Database (Neon DB via PostgreSQL Driver)
	config.ConnectDatabase()

	r := gin.Default()

	// 3. Setup CORS (Penting agar Next.js bisa akses)
	// Sesuaikan AllowOrigins dengan URL frontend Next.js Anda nanti
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	r.Use(cors.New(config))

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
	r.Run("0.0.0.0:" + port)
}
