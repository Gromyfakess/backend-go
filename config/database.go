package config

import (
	"fmt"
	"log"
	"os"
	"siro-backend/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	name := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, pass, host, port, name)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Gagal koneksi database: ", err)
	}

	// Tambahkan models.ActivityLog di sini
	err = DB.AutoMigrate(&models.User{}, &models.UserToken{}, &models.WorkOrder{}, &models.ActivityLog{})
	if err != nil {
		log.Fatal("Gagal migrasi database: ", err)
	}

	log.Println("Database Connected Successfully via .env config!")
}
