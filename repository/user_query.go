package repository

import (
	"siro-backend/config"
	"siro-backend/models"
)

func GetUserByEmail(email string) (models.User, error) {
	var user models.User
	err := config.DB.Where("email = ?", email).First(&user).Error
	return user, err
}

func GetUserByID(id uint) (models.User, error) {
	var user models.User
	err := config.DB.First(&user, id).Error
	return user, err
}

func GetAllUsers() []models.User {
	var users []models.User
	config.DB.Find(&users)
	return users
}

func CreateUser(user *models.User) error {
	return config.DB.Create(user).Error
}

func UpdateUser(user *models.User) error {
	return config.DB.Save(user).Error
}

func DeleteUser(id string) error {
	return config.DB.Delete(&models.User{}, id).Error
}

func GetStaffByUnit(unit string) []models.User {
	var users []models.User
	config.DB.Where("unit = ?", unit).Find(&users)
	return users
}

func UpdateAvailability(id string, status string) error {
	return config.DB.Model(&models.User{}).Where("id = ?", id).Update("availability", status).Error
}
