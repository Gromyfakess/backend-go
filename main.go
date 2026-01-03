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
	// 1. Load Environment Variables
	if err := godotenv.Load(); err != nil {
		log.Println("Info: .env file not found, using system environment variables")
	}

	// 2. Initialize System
	utils.InitJWT()
	config.ConnectDB()

	// 3. Setup Router
	r := gin.Default()

	// --- CORS CONFIG ---
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{frontendURL},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * 3600, // Cache preflight 12 jam
	}))

	r.Static("/uploads", "./uploads")

	// Setup Routes
	setupRoutes(r)

	// Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("🚀 Server running securely on port :%s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Server failed to start: ", err)
	}
}

// setupRoutes: Mengelompokkan route agar main() tetap bersih
func setupRoutes(r *gin.Engine) {
	// --- PUBLIC ROUTES ---
	r.POST("/login", controllers.LoginHandler)
	r.POST("/refresh", controllers.RefreshHandler)
	r.POST("/logout", controllers.LogoutHandler)

	// --- PROTECTED ROUTES ---
	api := r.Group("/")
	api.Use(middleware.AuthMiddleware())
	{
		// 1. User & Profile
		api.GET("/me", controllers.GetMe)
		api.PUT("/me", controllers.UpdateMe)
		api.POST("/upload", controllers.UploadFile)

		// Staff
		api.GET("/staff", controllers.GetStaffList)
		api.PATCH("/staff/:id/availability", controllers.UpdateAvailability)

		// Activity Logs
		api.GET("/activities", controllers.GetActivities)

		// Work Orders
		wo := api.Group("/workorders")
		{
			wo.GET("", controllers.GetWorkOrders)
			wo.POST("", controllers.CreateWorkOrder)

			// Action Routes
			wo.PATCH("/:id/take", controllers.TakeRequest)
			wo.PATCH("/:id/assign", controllers.AssignStaff)
			wo.PATCH("/:id/finalize", controllers.FinalizeOrder)
		}

		api.POST("/upload/workorder", controllers.UploadWorkOrderEvidence)

		// Admin Management
		admin := api.Group("/admin")
		admin.Use(middleware.AdminOnly())
		{
			admin.GET("/users", controllers.GetAllUsers)
			admin.POST("/users", controllers.CreateUser)
			admin.PUT("/users/:id", controllers.UpdateUser)
			admin.DELETE("/users/:id", controllers.DeleteUser)
		}
	}
}
