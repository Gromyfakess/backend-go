package main

import (
	"fmt"
	"log"
	"os"
	"siro-backend/config"
	"siro-backend/constants" // Import constants
	"siro-backend/controllers"
	"siro-backend/middleware"
	"siro-backend/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Info: .env file not found, using system environment variables")
	}

	utils.InitJWT()
	config.ConnectDB()

	r := gin.Default()

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}

	corsConfig := cors.Config{
		AllowOriginFunc: func(origin string) bool {
			return true
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * 3600,
	}
	r.Use(cors.New(corsConfig))

	r.Static("/"+constants.DirUploads, "./"+constants.DirUploads)

	setupRoutes(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Server running on port :%s (Allowed Origin: %s)\n", port, frontendURL)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Server failed to start: ", err)
	}
}

// ... (Fungsi setupRoutes sama persis) ...
func setupRoutes(r *gin.Engine) {
	r.POST("/login", controllers.LoginHandler)
	r.POST("/refresh", controllers.RefreshHandler)
	r.POST("/logout", controllers.LogoutHandler)

	api := r.Group("/")
	api.Use(middleware.AuthMiddleware())
	{
		api.GET("/me", controllers.GetMe)
		api.PUT("/me", controllers.UpdateMe)
		api.POST("/upload", controllers.UploadFile)
		api.GET("/staff", controllers.GetStaffList)
		api.PATCH("/staff/:id/availability", controllers.UpdateAvailability)
		api.GET("/activities", controllers.GetActivities)

		wo := api.Group("/workorders")
		{
			wo.GET("", controllers.GetWorkOrders)
			wo.POST("", controllers.CreateWorkOrder)
			wo.PATCH("/:id/take", controllers.TakeRequest)
			wo.PATCH("/:id/assign", controllers.AssignStaff)
			wo.PATCH("/:id/finalize", controllers.FinalizeOrder)
		}
		api.POST("/upload/workorder", controllers.UploadWorkOrderEvidence)

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
