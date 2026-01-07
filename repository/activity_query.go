package repository

import (
	"fmt"
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

func GetActivities() ([]models.ActivityLog, error) {
	query := `SELECT id, user_id, user_name, action, request_id, details, status, timestamp 
              FROM activity_logs ORDER BY timestamp DESC LIMIT 50`

	rows, err := config.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.ActivityLog
	for rows.Next() {
		var l models.ActivityLog
		err := rows.Scan(
			&l.ID, &l.UserID, &l.UserName, &l.Action, &l.RequestID,
			&l.Details, &l.Status, &l.Timestamp,
		)
		if err == nil {
			logs = append(logs, l)
		}
	}
	return logs, nil
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
