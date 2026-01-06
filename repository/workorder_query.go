package repository

import (
	"siro-backend/config"
	"siro-backend/constants"
	"siro-backend/models"
	"time"
)

func CreateWorkOrder(wo *models.WorkOrder) error {
	return config.DB.Create(wo).Error
}

func GetWorkOrderById(id uint) (models.WorkOrder, error) {
	var wo models.WorkOrder
	err := config.DB.Preload("Assignee").
		Preload("RequesterData").
		Preload("CompletedBy").
		First(&wo, id).Error
	return wo, err
}

func GetWorkOrders(filters map[string]string) []models.WorkOrder {
	var wos []models.WorkOrder

	query := config.DB.Model(&models.WorkOrder{}).
		Preload("Assignee").
		Preload("RequesterData").
		Preload("CompletedBy")

	// Filter Status
	if status, ok := filters["status"]; ok && status != "" {
		if status == "active" {
			query = query.Where("work_orders.status IN ?", []string{constants.StatusPending, constants.StatusInProgress})
		} else {
			query = query.Where("work_orders.status = ?", status)
		}
	}

	// Filter Unit Tujuan (Incoming Requests)
	if unit, ok := filters["unit"]; ok && unit != "" {
		query = query.Where("work_orders.unit = ?", unit)
	}

	// Filter Unit Pembuat (Outgoing Requests)
	if reqUnit, ok := filters["requester_unit"]; ok && reqUnit != "" {
		query = query.Joins("JOIN users ON users.id = work_orders.requester_id").
			Where("users.unit = ?", reqUnit)
	}

	// Filter Waktu (Today)
	if date, ok := filters["date"]; ok && date == "today" {
		now := time.Now()
		// Set waktu start ke 00:00:00 hari ini
		startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		// Set waktu end ke 00:00:00 besok
		endOfDay := startOfDay.Add(24 * time.Hour)

		query = query.Where("work_orders.created_at >= ? AND work_orders.created_at < ?", startOfDay, endOfDay)
	}

	query.Order("work_orders.created_at desc").Limit(20).Find(&wos)
	return wos
}

func UpdateWorkOrder(wo *models.WorkOrder) error {
	return config.DB.Save(wo).Error
}

// -- ini jika ingin buat fungsi update dan delete (sekarang blm perlu) --

// func DeleteWorkOrder(wo *models.WorkOrder) error {
// 	return config.DB.Delete(wo).Error
// }
