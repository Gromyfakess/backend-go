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

func ConnectDB() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=sirodb port=5432 sslmode=disable TimeZone=Asia/Jakarta"
	}

	// 1. Buka Koneksi (Ini biasanya cepat)
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		// Opsional: SkipDefaultTransaction bisa mempercepat insert masif
		SkipDefaultTransaction: true,
	})

	if err != nil {
		panic("Failed to connect to database!")
	}

	DB = database
	fmt.Println("✅ Connected to Neon DB (PostgreSQL)")

	// 2. Jalankan Migrasi di BACKGROUND (Goroutine)
	// Agar tidak memblokir startup server HTTP
	go func() {
		fmt.Println("⏳ Starting AutoMigrate in background...")
		err = database.AutoMigrate(
			&models.User{},
			&models.WorkOrder{},
			&models.ActivityLog{},
		)
		if err != nil {
			log.Println("❌ Migration failed:", err)
		} else {
			fmt.Println("✅ AutoMigrate finished successfully")
		}
	}()
}
