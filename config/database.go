package config

import (
	"fmt"
	"log"
	"os"
	"siro-backend/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	// Ambil koneksi string dari Environment Variable (Diset di Dashboard Leapcell)
	dsn := os.Getenv("DATABASE_URL")

	if dsn == "" {
		// Fallback untuk local dev jika env file belum diload atau kosong
		dsn = "host=localhost user=postgres password=postgres dbname=sirodb port=5432 sslmode=disable TimeZone=Asia/Jakarta"
		fmt.Println("⚠️  Warning: DATABASE_URL not set, using default local config")
	}

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Gagal koneksi database: ", err)
	}

	// Auto Migrate
	err = database.AutoMigrate(
		&models.User{},
		&models.WorkOrder{},
		&models.ActivityLog{},
	)

	if err != nil {
		log.Fatal("Migration failed:", err)
	}

	DB = database
	fmt.Println("✅ Connected to Neon DB (PostgreSQL)")
}
