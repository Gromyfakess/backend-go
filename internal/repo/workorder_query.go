package repo

import (
	"database/sql"
	"fmt"
	"math"
	"siro-backend/global"
	"siro-backend/internal/models"
	"siro-backend/pkg/setting"
	"strings"
)

const selectWOQuery = `
    SELECT 
        w.id, w.title, w.description, w.priority, w.status, w.unit, w.photo_url, 
        w.requester_id, w.assignee_id, w.taken_at, 
        w.completed_at, w.completed_by_id, COALESCE(w.completion_note, ''), w.created_at, w.updated_at,
        req.name, req.unit, COALESCE(req.avatar_url, ''),     	 				  -- Requester Info
        COALESCE(asg.name, ''), COALESCE(asg.email, ''), COALESCE(asg.unit, ''),  -- Assignee Info
        COALESCE(cmp.name, '')                                  				  -- CompletedBy Info
    FROM work_orders w
    LEFT JOIN users req ON w.requester_id = req.id
    LEFT JOIN users asg ON w.assignee_id = asg.id
    LEFT JOIN users cmp ON w.completed_by_id = cmp.id
`

// Helper Scan (INI YANG DIPERBAIKI)
func scanWO(rows *sql.Rows) (models.WorkOrder, error) {
	var w models.WorkOrder
	var asgID, cmpID sql.NullInt64
	var takenAt, completedAt sql.NullTime

	err := rows.Scan(
		&w.ID, &w.Title, &w.Description, &w.Priority, &w.Status, &w.Unit, &w.PhotoURL,
		&w.RequesterID, &asgID, &takenAt, &completedAt, &cmpID, &w.CompletionNote, &w.CreatedAt, &w.UpdatedAt,
		&w.RequesterData.Name, &w.RequesterData.Unit, &w.RequesterData.AvatarURL,
		&w.Assignee.Name, &w.Assignee.Email, &w.Assignee.Unit,
		&w.CompletedBy.Name,
	)
	if err != nil {
		return w, err
	}

	// --- FIX: Mapping Nama Flat String ---
	// Frontend butuh field 'requester' (flat string), bukan cuma 'requesterData.name'
	w.RequesterName = w.RequesterData.Name
	// -------------------------------------

	if asgID.Valid {
		uid := uint(asgID.Int64)
		w.AssigneeID = &uid
		w.Assignee.ID = uid
	}
	if cmpID.Valid {
		uid := uint(cmpID.Int64)
		w.CompletedByID = &uid
		w.CompletedBy.ID = uid
	}
	if takenAt.Valid {
		w.TakenAt = &takenAt.Time
	}
	if completedAt.Valid {
		w.CompletedAt = &completedAt.Time
	}

	return w, nil
}

func GetDashboardStats(userUnit string) (models.DashboardStats, error) {
	var stats models.DashboardStats

	// 1. Hitung Incoming (Total, Pending, In Progress) untuk Unit Saya
	// Menggunakan Conditional Sum agar hanya 1x query ke DB
	queryIncoming := `
		SELECT 
			COUNT(*) as total,
			COALESCE(SUM(CASE WHEN status = ? THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN status = ? THEN 1 ELSE 0 END), 0)
		FROM work_orders 
		WHERE unit = ?`

	err := setting.DB.QueryRow(queryIncoming, global.StatusPending, global.StatusInProgress, userUnit).
		Scan(&stats.Incoming, &stats.Pending, &stats.InProgress)

	if err != nil {
		return stats, err
	}

	// 2. Hitung Outgoing (Tiket yang dibuat oleh Unit Saya ke unit lain)
	// Membutuhkan JOIN ke tabel users untuk cek unit requester
	queryOutgoing := `
		SELECT COUNT(*) 
		FROM work_orders w
		JOIN users req ON w.requester_id = req.id
		WHERE req.unit = ?`

	err = setting.DB.QueryRow(queryOutgoing, userUnit).Scan(&stats.Outgoing)

	return stats, err
}

func CreateWorkOrder(wo *models.WorkOrder) error {
	query := `INSERT INTO work_orders (title, description, priority, status, unit, photo_url, requester_id, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`
	res, err := setting.DB.Exec(query, wo.Title, wo.Description, wo.Priority, global.StatusPending, wo.Unit, wo.PhotoURL, wo.RequesterID)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	wo.ID = uint(id)
	return nil
}

func GetWorkOrderById(id uint) (models.WorkOrder, error) {
	rows, err := setting.DB.Query(selectWOQuery+" WHERE w.id = ?", id)
	if err != nil {
		return models.WorkOrder{}, err
	}
	defer rows.Close()
	if rows.Next() {
		return scanWO(rows)
	}
	return models.WorkOrder{}, fmt.Errorf("not found")
}

func GetWorkOrders(filters map[string]string, page, limit int) ([]models.WorkOrder, models.PaginationMeta, error) {
	query := selectWOQuery
	countQuery := "SELECT COUNT(*) FROM work_orders w LEFT JOIN users req ON w.requester_id = req.id"

	var conditions []string
	var args []interface{}

	if s := filters["status"]; s != "" {
		if s == "active" {
			conditions = append(conditions, "w.status IN (?, ?)")
			args = append(args, "Pending", "In Progress")
		} else {
			conditions = append(conditions, "w.status = ?")
			args = append(args, s)
		}
	}
	if u := filters["unit"]; u != "" {
		conditions = append(conditions, "w.unit = ?")
		args = append(args, u)
	}
	if ru := filters["requester_unit"]; ru != "" {
		conditions = append(conditions, "req.unit = ?")
		args = append(args, ru)
	}
	if filters["date"] == "today" {
		conditions = append(conditions, "DATE(w.created_at) = CURDATE()")
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	var totalItems int
	err := setting.DB.QueryRow(countQuery+whereClause, args...).Scan(&totalItems)
	if err != nil {
		return nil, models.PaginationMeta{}, err
	}

	offset := (page - 1) * limit
	query += whereClause + " ORDER BY w.created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := setting.DB.Query(query, args...)
	if err != nil {
		return nil, models.PaginationMeta{}, err
	}
	defer rows.Close()

	var wos []models.WorkOrder
	for rows.Next() {
		if wo, err := scanWO(rows); err == nil {
			wos = append(wos, wo)
		}
	}

	totalPages := int(math.Ceil(float64(totalItems) / float64(limit)))
	meta := models.PaginationMeta{
		CurrentPage: page,
		TotalPages:  totalPages,
		TotalItems:  totalItems,
		Limit:       limit,
	}

	return wos, meta, nil
}

func TakeWorkOrder(woID, userID uint) error {
	res, err := setting.DB.Exec("UPDATE work_orders SET status=?, assignee_id=?, taken_at=NOW(), updated_at=NOW() WHERE id=? AND assignee_id IS NULL",
		global.StatusInProgress, userID, woID)
	if err != nil {
		return err
	}
	if aff, _ := res.RowsAffected(); aff == 0 {
		return fmt.Errorf("tiket sudah diambil")
	}
	return nil
}

func AssignWorkOrder(woID, userID uint) error {
	_, err := setting.DB.Exec("UPDATE work_orders SET status=?, assignee_id=?, updated_at=NOW() WHERE id=?",
		global.StatusInProgress, userID, woID)
	return err
}

func FinalizeWorkOrder(woID uint, note string, userID uint) error {
	_, err := setting.DB.Exec("UPDATE work_orders SET status=?, completion_note=?, completed_at=NOW(), completed_by_id=?, updated_at=NOW() WHERE id=?",
		global.StatusCompleted, note, userID, woID)
	return err
}
