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
	err := config.DB.Preload("Assignee").
		Preload("RequesterData").
		Preload("CompletedBy").
		First(&wo, id).Error
	return wo, err
}

func GetAllWorkOrders() []models.WorkOrder {
	var wos []models.WorkOrder
	config.DB.Preload("Assignee").
		Preload("RequesterData").
		Preload("CompletedBy").
		Order("created_at desc").
		Find(&wos)
	return wos
}

func UpdateWorkOrder(wo *models.WorkOrder) error {
	return config.DB.Save(wo).Error
}

func DeleteWorkOrder(wo *models.WorkOrder) error {
	return config.DB.Delete(wo).Error
}
