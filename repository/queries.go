package repository

import (
	"siro-backend/config"
	"siro-backend/models"
	"time"

	"gorm.io/gorm/clause"
)

// --- USER QUERIES ---

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

// --- TOKEN QUERIES (LOGIC SINGLE SESSION) ---

// SaveAllTokens: Upsert (Update or Insert) semua token
func SaveAllTokens(userID uint, access, refresh string, atExp, rtExp time.Time) error {
	tokenData := models.UserToken{
		UserID:       userID,
		AccessToken:  access,
		RefreshToken: refresh,
		ATExpiresAt:  atExp,
		RTExpiresAt:  rtExp,
	}
	// OnConflict: Jika UserID sudah ada, update tokennya.
	return config.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"access_token", "refresh_token", "at_expires_at", "rt_expires_at"}),
	}).Create(&tokenData).Error
}

// UpdateAccessTokenOnly: Hanya update access token, refresh token dibiarkan
func UpdateAccessTokenOnly(userID uint, newAccess string, newAtExp time.Time) error {
	return config.DB.Model(&models.UserToken{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"access_token":  newAccess,
			"at_expires_at": newAtExp,
		}).Error
}

// CheckRefreshTokenValid: Cek validitas & ambil info expiry
func CheckRefreshTokenValid(userID uint, refreshString string) (bool, time.Time) {
	var tokenData models.UserToken
	err := config.DB.Where("user_id = ?", userID).First(&tokenData).Error
	if err != nil {
		return false, time.Time{}
	}
	// Token harus sama persis DAN belum expired
	isValid := (tokenData.RefreshToken == refreshString) && (time.Now().Before(tokenData.RTExpiresAt))
	return isValid, tokenData.RTExpiresAt
}

// CheckAccessTokenValid: Digunakan Middleware untuk cek token aktif
func CheckAccessTokenValid(userID uint, tokenString string) bool {
	var tokenData models.UserToken
	err := config.DB.Where("user_id = ?", userID).First(&tokenData).Error
	if err != nil {
		return false
	}
	return tokenData.AccessToken == tokenString
}

// --- WORK ORDER QUERIES ---

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
