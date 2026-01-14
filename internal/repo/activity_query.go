package repo

import (
	"fmt"
	"math"
	"siro-backend/internal/models"
	"siro-backend/pkg/setting"
	"time"
)

func CreateActivityLog(log models.ActivityLog) error {
	query := `INSERT INTO activity_logs (user_id, user_name, action, request_id, details, status, timestamp)
              VALUES (?, ?, ?, ?, ?, ?, NOW())`

	_, err := setting.DB.Exec(query, log.UserID, log.UserName, log.Action, log.RequestID, log.Details, log.Status)
	return err
}

// GetActivities returns paginated activity logs filtered by user's unit
// Only shows activities where the user's unit is involved (as requester unit OR target unit)
func GetActivities(userUnit string, page, limit int) ([]models.ActivityLog, models.PaginationMeta, error) {
	// Base query with JOIN to work_orders to filter by unit
	// Show activities where:
	// 1. The work order's target unit matches user's unit, OR
	// 2. The requester's unit matches user's unit
	baseQuery := `
		SELECT DISTINCT a.id, a.user_id, a.user_name, a.action, a.request_id, a.details, a.status, a.timestamp
		FROM activity_logs a
		INNER JOIN work_orders w ON a.request_id = w.id
		LEFT JOIN users req ON w.requester_id = req.id
		WHERE w.unit = ? OR req.unit = ?
	`

	// Count total matching activities
	countQuery := `
		SELECT COUNT(DISTINCT a.id)
		FROM activity_logs a
		INNER JOIN work_orders w ON a.request_id = w.id
		LEFT JOIN users req ON w.requester_id = req.id
		WHERE w.unit = ? OR req.unit = ?
	`

	var totalItems int
	err := setting.DB.QueryRow(countQuery, userUnit, userUnit).Scan(&totalItems)
	if err != nil {
		return nil, models.PaginationMeta{}, err
	}

	// Get paginated data
	offset := (page - 1) * limit
	query := baseQuery + " ORDER BY a.timestamp DESC LIMIT ? OFFSET ?"

	rows, err := setting.DB.Query(query, userUnit, userUnit, limit, offset)
	if err != nil {
		return nil, models.PaginationMeta{}, err
	}
	defer rows.Close()

	var logs []models.ActivityLog
	for rows.Next() {
		var l models.ActivityLog
		if err := rows.Scan(&l.ID, &l.UserID, &l.UserName, &l.Action, &l.RequestID, &l.Details, &l.Status, &l.Timestamp); err == nil {
			logs = append(logs, l)
		}
	}

	// Calculate pagination metadata
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
