package setting

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// DB is the database connection that other packages can use
var DB *sql.DB

// ConnectDB connects to MySQL database
// Reads connection info from environment variables
func ConnectDB() {
	// Get database settings from environment variables
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	name := os.Getenv("DB_NAME")

	// Check if all required settings are provided
	if user == "" || pass == "" || host == "" || port == "" || name == "" {
		log.Fatal("ERROR: Missing database configuration in .env file. Please check DB_USER, DB_PASSWORD, DB_HOST, DB_PORT, DB_NAME")
	}

	// Create connection string
	// parseTime=true is needed so MySQL timestamps work with Go's time.Time
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		user, pass, host, port, name)

	// Open database connection
	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("ERROR: Failed to open database connection: ", err)
	}

	// Configure connection pool for better performance
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(5 * time.Minute)

	// Test the connection by pinging the database
	if err := DB.Ping(); err != nil {
		log.Fatal("ERROR: Failed to connect to database. Please check your database settings: ", err)
	}

	log.Println("Database Connected Successfully!")
}
