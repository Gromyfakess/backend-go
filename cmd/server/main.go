package main

import (
	"fmt"
	"log"
	"os"
	"siro-backend/internal/initialize"
	"siro-backend/internal/routers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load env
	if err := godotenv.Load(); err != nil {
		log.Println("Info: .env file not found, using system environment variables")
	}

	// Initialize database and JWT
	initialize.Initialize()

	// Create router
	r := gin.Default()

	// Get frontend URL from env (for CORS)
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}

	// CORS
	corsConfig := cors.Config{
		AllowOriginFunc: func(origin string) bool {
			return origin == frontendURL || origin == "http://localhost:3000"
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * 3600, // Cache preflight requests for 12 hours
	}
	r.Use(cors.New(corsConfig))

	// Setup all routes
	routers.SetupRoutes(r)

	// Get port from env
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Get server host from environment
	// Change to 0.0.0.0 in .env if you need network access
	host := os.Getenv("SERVER_HOST")
	if host == "" {
		host = "localhost"
	}

	// Create address string
	address := fmt.Sprintf("%s:%s", host, port)

	// Show startup message
	if host == "localhost" {
		fmt.Printf("Server running on http://localhost:%s\n", port)
		fmt.Printf("Frontend URL: %s\n", frontendURL)
		fmt.Printf("(Only accessible from this computer)\n")
	} else {
		fmt.Printf("Server running on %s\n", address)
		fmt.Printf("Frontend URL: %s\n", frontendURL)
		fmt.Printf("(Accessible from network)\n")
	}

	// Start server
	if err := r.Run(address); err != nil {
		log.Fatal("Server failed to start: ", err)
	}
}
