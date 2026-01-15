package routers

import (
	"siro-backend/global"
	"siro-backend/internal/controller"
	"siro-backend/internal/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	r.Static("/"+global.DirUploads, "./"+global.DirUploads)

	r.POST("/login", controller.LoginHandler)
	r.POST("/refresh", controller.RefreshHandler)

	api := r.Group("/")

	api.Use(middlewares.AuthMiddleware())

	{
		api.POST("/logout", controller.LogoutHandler)
		api.GET("/me", controller.GetMe)
		api.PUT("/me", controller.UpdateMe)
		api.POST("/upload", controller.UploadFile)
		api.GET("/staff", controller.GetStaffList)
		api.PATCH("/staff/:id/availability", controller.UpdateAvailability)
		api.GET("/activities", controller.GetActivities)

		wo := api.Group("/workorders")
		{
			wo.GET("/stats", controller.GetStats)
			wo.GET("", controller.GetWorkOrders)
			wo.POST("", controller.CreateWorkOrder)
			wo.PATCH("/:id/take", controller.TakeRequest)
			wo.PATCH("/:id/assign", controller.AssignStaff)
			wo.PATCH("/:id/finalize", controller.FinalizeOrder)
		}
		api.POST("/upload/workorder", controller.UploadWorkOrderEvidence)

		admin := api.Group("/admin")
		admin.Use(middlewares.AdminOnly())
		{
			admin.GET("/users", controller.GetAllUsers)
			admin.POST("/users", controller.CreateUser)
			admin.PUT("/users/:id", controller.UpdateUser)
			admin.DELETE("/users/:id", controller.DeleteUser)
		}
	}
}
