package models

import "time"

// User
type User struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name"`
	Email        string    `json:"email" gorm:"unique;size:100"`
	Role         string    `json:"role"`
	Unit         string    `json:"unit"`
	Phone        string    `json:"phone"`
	AvatarURL    string    `json:"avatar"`
	Availability string    `json:"availability"`
	CanCRUD      bool      `json:"canCRUD"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"-"`
}

// UserToken: Tabel sesi
type UserToken struct {
	UserID       uint   `gorm:"primaryKey"`
	AccessToken  string `gorm:"type:text"`
	RefreshToken string `gorm:"type:text"`
	ATExpiresAt  time.Time
	RTExpiresAt  time.Time
}

// WorkOrder: Tabel request/tiket
type WorkOrder struct {
	ID            uint   `json:"id" gorm:"primaryKey"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	Priority      string `json:"priority"`
	Status        string `json:"status"`
	Unit          string `json:"unit"`
	RequesterID   uint   `json:"requesterId"`
	RequesterName string `json:"requester"`
	// Relasi ke User
	RequesterData User `json:"requesterData" gorm:"foreignKey:RequesterID"`

	AssigneeID *uint `json:"assigneeId"`
	Assignee   User  `json:"assignee" gorm:"foreignKey:AssigneeID"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// --- Request Structs ---

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type WorkOrderRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Priority    string `json:"priority" binding:"required"`
	Unit        string `json:"unit"`
}

type UserRequest struct {
	Name      string `json:"name" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password"`
	Role      string `json:"role" binding:"required"`
	Unit      string `json:"unit" binding:"required"`
	Phone     string `json:"phone"`
	CanCRUD   bool   `json:"canCRUD"`
	AvatarURL string `json:"avatar"`
}

type AssignRequest struct {
	AssigneeID uint `json:"assigneeId" binding:"required"`
}

type AvailabilityRequest struct {
	Status string `json:"status" binding:"required"`
}
