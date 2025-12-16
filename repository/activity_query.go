package repository

import (
	"siro-backend/config"
	"siro-backend/models"
)

func CreateActivityLog(log *models.ActivityLog) {
	// Gunakan goroutine agar tidak memblokir response utama
	go func() {
		config.DB.Create(log)
	}()
}

func GetRecentActivities() []models.ActivityLog {
	var logs []models.ActivityLog
	// Ambil 5 aktivitas terakhir
	config.DB.Order("timestamp desc").Limit(5).Find(&logs)
	return logs
}
