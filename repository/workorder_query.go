package repository

import (
	"siro-backend/config"
	"siro-backend/models"
)

func CreateWorkOrder(wo *models.WorkOrder) error {
	return config.DB.Create(wo).Error
}

func GetWorkOrderById(id string) (models.WorkOrder, error) {
	var wo models.WorkOrder
	// Preload Assignee DAN RequesterData (untuk ambil avatar user)
	err := config.DB.Preload("Assignee").Preload("RequesterData").First(&wo, id).Error
	return wo, err
}

func GetAllWorkOrders() []models.WorkOrder {
	var wos []models.WorkOrder
	// Preload Assignee DAN RequesterData
	config.DB.Preload("Assignee").Preload("RequesterData").Order("created_at desc").Find(&wos)
	return wos
}

func UpdateWorkOrder(wo *models.WorkOrder) error {
	return config.DB.Save(wo).Error
}

func DeleteWorkOrder(wo *models.WorkOrder) error {
	return config.DB.Delete(wo).Error
}
