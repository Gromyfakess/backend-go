package repository

import (
	"siro-backend/config"
	"siro-backend/models"
)

// CreateActivityLog: Menyimpan log aktivitas
func CreateActivityLog(log *models.ActivityLog) error {
	return config.DB.Create(log).Error
}

// GetRecentActivities: Mengambil log aktivitas terbaru dengan limit tertentu
func GetRecentActivities(limit int) []models.ActivityLog {
	var logs []models.ActivityLog

	if limit <= 0 {
		limit = 5
	}

	config.DB.Order("timestamp desc").Limit(limit).Find(&logs)
	return logs
}
