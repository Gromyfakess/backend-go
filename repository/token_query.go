package repository

import (
	"siro-backend/config"
	"siro-backend/models"
	"time"

	"gorm.io/gorm/clause"
)

// SaveAllTokens: Upsert access & refresh token
func SaveAllTokens(userID uint, access, refresh string, atExp, rtExp time.Time) error {
	tokenData := models.UserToken{
		UserID:       userID,
		AccessToken:  access,
		RefreshToken: refresh,
		ATExpiresAt:  atExp,
		RTExpiresAt:  rtExp,
	}

	return config.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"access_token", "refresh_token", "at_expires_at", "rt_expires_at"}),
	}).Create(&tokenData).Error
}

// UpdateAccessTokenOnly: Hanya rotasi access token (digunakan saat refresh token masih valid)
func UpdateAccessTokenOnly(userID uint, newAccess string, newAtExp time.Time) error {
	return config.DB.Model(&models.UserToken{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"access_token":  newAccess,
			"at_expires_at": newAtExp,
		}).Error
}

// CheckRefreshTokenValid: Memeriksa apakah refresh token di database cocok dan belum expired
func CheckRefreshTokenValid(userID uint, refreshString string) (bool, time.Time) {
	var tokenData models.UserToken

	// Fail-fast jika record tidak ditemukan
	if err := config.DB.Where("user_id = ?", userID).First(&tokenData).Error; err != nil {
		return false, time.Time{}
	}

	isTokenMatch := tokenData.RefreshToken == refreshString
	isNotExpired := time.Now().Before(tokenData.RTExpiresAt)

	return (isTokenMatch && isNotExpired), tokenData.RTExpiresAt
}

// CheckAccessTokenValid: Validasi sederhana untuk middleware
func CheckAccessTokenValid(userID uint, tokenString string) bool {
	var tokenData models.UserToken

	if err := config.DB.Select("access_token").Where("user_id = ?", userID).First(&tokenData).Error; err != nil {
		return false
	}

	return tokenData.AccessToken == tokenString
}
