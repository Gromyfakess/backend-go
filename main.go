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

		api.GET("/staff", controllers.GetStaffList)
		api.PATCH("/staff/:id/availability", controllers.UpdateAvailability)

		// User biasa CRUD workorders (tetap di root /workorders)
		api.GET("/workorders", controllers.GetWorkOrders)
		api.POST("/workorders", controllers.CreateWorkOrder)
		api.PUT("/workorders/:id", controllers.UpdateWorkOrder)
		api.DELETE("/workorders/:id", controllers.DeleteWorkOrder)

		api.POST("/upload/workorder", controllers.UploadWorkOrderEvidence)

		api.PATCH("/workorders/:id/take", controllers.TakeRequest)
		api.PATCH("/workorders/:id/assign", controllers.AssignStaff)
		api.PATCH("/workorders/:id/finalize", controllers.FinalizeOrder)

		api.GET("/activities", controllers.GetActivities)

		// === ADMIN ROUTES ===
		// Semua rute admin dikelompokkan di bawah /admin
		admin := api.Group("/admin")
		admin.Use(middleware.AdminOnly())
		{
			// Manajemen User
			admin.GET("/users", controllers.GetAllUsers)       // GET /admin/users
			admin.POST("/users", controllers.CreateUser)       // POST /admin/users
			admin.PUT("/users/:id", controllers.UpdateUser)    // PUT /admin/users/:id
			admin.DELETE("/users/:id", controllers.DeleteUser) // DELETE /admin/users/:id

			// Manajemen Workorder Admin (Opsional, jika ingin memisahkan view admin)
			admin.GET("/workorders", controllers.GetWorkOrders)
			admin.POST("/workorders", controllers.CreateWorkOrder)
			admin.PUT("/workorders/:id", controllers.UpdateWorkOrder)
			admin.DELETE("/workorders/:id", controllers.DeleteWorkOrder)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Server running on port :%s\n", port)
	r.Run(":" + port)
}
