package repository

import (
	"database/sql"
	"siro-backend/config"
	"time"
)

// SaveToken: Menyimpan token baru atau mengupdate yang lama (Login/Register)
func SaveToken(userID uint, access, refresh string, atExp, rtExp time.Time) error {
	// Fitur spesial MySQL: Insert Or Update (Upsert)
	query := `INSERT INTO user_tokens (user_id, access_token, refresh_token, at_expires_at, rt_expires_at)
			  VALUES (?, ?, ?, ?, ?)
			  ON DUPLICATE KEY UPDATE 
			  access_token = VALUES(access_token), 
			  refresh_token = VALUES(refresh_token),
			  at_expires_at = VALUES(at_expires_at), 
			  rt_expires_at = VALUES(rt_expires_at)`

	_, err := config.DB.Exec(query, userID, access, refresh, atExp, rtExp)
	return err
}

// UpdateAccessTokenOnly: Hanya rotasi access token baru (digunakan saat Refresh Token)
func UpdateAccessTokenOnly(userID uint, newAccess string, newAtExp time.Time) error {
	query := `UPDATE user_tokens SET access_token = ?, at_expires_at = ? WHERE user_id = ?`

	_, err := config.DB.Exec(query, newAccess, newAtExp, userID)
	return err
}

// CheckRefreshTokenValid: Memeriksa apakah refresh token valid dan belum expired
func CheckRefreshTokenValid(userID uint, refreshString string) (bool, time.Time) {
	var dbRefreshToken string
	var rtExpiresAt time.Time

	// Ambil refresh token & expiry dari DB
	query := `SELECT refresh_token, rt_expires_at FROM user_tokens WHERE user_id = ?`

	err := config.DB.QueryRow(query, userID).Scan(&dbRefreshToken, &rtExpiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, time.Time{} // User tidak punya token session
		}
		return false, time.Time{} // Error lain
	}

	// Logic Validasi
	isTokenMatch := (dbRefreshToken == refreshString)
	isNotExpired := time.Now().Before(rtExpiresAt)

	return (isTokenMatch && isNotExpired), rtExpiresAt
}

// CheckAccessTokenValid: (Opsional) Validasi tambahan untuk middleware jika diperlukan
// Berguna untuk fitur "Force Logout" (mendeteksi jika token di DB sudah berubah/dihapus)
func CheckAccessTokenValid(userID uint, tokenString string) bool {
	var dbAccessToken string

	query := `SELECT access_token FROM user_tokens WHERE user_id = ?`

	err := config.DB.QueryRow(query, userID).Scan(&dbAccessToken)
	if err != nil {
		return false // Token tidak ditemukan
	}

	return dbAccessToken == tokenString
}

// DeleteToken: Untuk Logout
func DeleteToken(userID uint) error {
	query := `DELETE FROM user_tokens WHERE user_id = ?`
	_, err := config.DB.Exec(query, userID)
	return err
}
