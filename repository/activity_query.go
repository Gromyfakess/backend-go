package repository

import (
	"fmt"
	"math"
	"siro-backend/config"
	"siro-backend/models"
	"time"
)

func CreateActivityLog(log models.ActivityLog) error {
	query := `INSERT INTO activity_logs (user_id, user_name, action, request_id, details, status, timestamp)
              VALUES (?, ?, ?, ?, ?, ?, NOW())`

	_, err := config.DB.Exec(query, log.UserID, log.UserName, log.Action, log.RequestID, log.Details, log.Status)
	return err
}

func GetActivities(page, limit int) ([]models.ActivityLog, models.PaginationMeta, error) {
	// 1. Hitung Total
	var totalItems int
	config.DB.QueryRow("SELECT COUNT(*) FROM activity_logs").Scan(&totalItems)

	// 2. Ambil Data
	offset := (page - 1) * limit
	query := `SELECT id, user_id, user_name, action, request_id, details, status, timestamp 
              FROM activity_logs ORDER BY timestamp DESC LIMIT ? OFFSET ?`

	rows, err := config.DB.Query(query, limit, offset)
	if err != nil {
		return nil, models.PaginationMeta{}, err
	}
	defer rows.Close()

	var logs []models.ActivityLog
	for rows.Next() {
		var l models.ActivityLog
		err := rows.Scan(&l.ID, &l.UserID, &l.UserName, &l.Action, &l.RequestID, &l.Details, &l.Status, &l.Timestamp)
		if err == nil {
			logs = append(logs, l)
		}
	}

	// 3. Meta
	totalPages := int(math.Ceil(float64(totalItems) / float64(limit)))
	meta := models.PaginationMeta{
		CurrentPage: page,
		TotalPages:  totalPages,
		TotalItems:  totalItems,
		Limit:       limit,
	}

	return logs, meta, nil
}

// LogActivity: Global async logger helper
func LogActivity(userID uint, userName, action, details, status string, reqID uint) {
	go func() {
		newLog := models.ActivityLog{
			UserID:    userID,
			UserName:  userName,
			Action:    action,
			Details:   details,
			Status:    status,
			RequestID: reqID,
			Timestamp: time.Now(),
		}
		if err := CreateActivityLog(newLog); err != nil {
			fmt.Printf("[LOG ERROR] Failed to save activity: %v\n", err)
		}
	}()
}
