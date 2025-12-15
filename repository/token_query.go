package repository

import (
	"siro-backend/config"
	"siro-backend/models"
	"time"

	"gorm.io/gorm/clause"
)

// SaveAllTokens: Update / Insert semua token
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
	// Token harus sama persis & belum expired
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
