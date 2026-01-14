package repo

import (
	"siro-backend/internal/models"
	"siro-backend/pkg/setting"
)

func GetUserByEmail(email string) (*models.User, error) {
	query := `SELECT id, name, email, password_hash, role, unit, availability, can_crud, COALESCE(avatar_url, '') 
              FROM users WHERE email = ?`
	var u models.User
	err := setting.DB.QueryRow(query, email).Scan(
		&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.Unit, &u.Availability, &u.CanCRUD, &u.AvatarURL,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func GetUserByID(id uint) (*models.User, error) {
	query := `SELECT id, name, email, role, unit, COALESCE(phone, ''), COALESCE(avatar_url, ''), availability, can_crud 
              FROM users WHERE id = ?`
	var u models.User
	err := setting.DB.QueryRow(query, id).Scan(
		&u.ID, &u.Name, &u.Email, &u.Role, &u.Unit, &u.Phone, &u.AvatarURL, &u.Availability, &u.CanCRUD,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func CreateUser(u *models.User) error {
	query := `INSERT INTO users (name, email, password_hash, role, unit, phone, can_crud, availability, avatar_url, created_at) 
              VALUES (?, ?, ?, ?, ?, ?, ?, 'Online', ?, NOW())`
	res, err := setting.DB.Exec(query, u.Name, u.Email, u.PasswordHash, u.Role, u.Unit, u.Phone, u.CanCRUD, u.AvatarURL)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	u.ID = uint(id)
	return nil
}

func GetAllUsers() ([]models.User, error) {
	rows, err := setting.DB.Query("SELECT id, name, email, role, unit, availability, COALESCE(avatar_url, '') FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.Unit, &u.Availability, &u.AvatarURL); err == nil {
			users = append(users, u)
		}
	}
	return users, nil
}

// GetUsersByUnit: Filter langsung di DB (Optimasi RAM & Performance)
func GetUsersByUnit(unit string) ([]models.User, error) {
	query := `SELECT id, name, email, role, unit, availability, COALESCE(avatar_url, '') 
	          FROM users WHERE unit = ?`

	rows, err := setting.DB.Query(query, unit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.Unit, &u.Availability, &u.AvatarURL); err == nil {
			users = append(users, u)
		}
	}
	return users, nil
}

func UpdateUser(id uint, u models.User) error {
	query := `UPDATE users SET name=?, unit=?, phone=?, role=?, can_crud=?, avatar_url=? WHERE id=?`
	_, err := setting.DB.Exec(query, u.Name, u.Unit, u.Phone, u.Role, u.CanCRUD, u.AvatarURL, id)
	return err
}

func UpdateAvailability(userID uint, status string) error {
	_, err := setting.DB.Exec("UPDATE users SET availability = ? WHERE id = ?", status, userID)
	return err
}

func DeleteUser(id uint) error {
	_, err := setting.DB.Exec("DELETE FROM users WHERE id = ?", id)
	return err
}
