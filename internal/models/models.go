package models

import "time"

// User
type User struct {
	ID           uint      `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Role         string    `json:"role"`
	Unit         string    `json:"unit"`
	Phone        string    `json:"phone"`
	AvatarURL    string    `json:"avatar"`
	Availability string    `json:"availability"`
	CanCRUD      bool      `json:"canCRUD"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"-"`
}

// UserToken
type UserToken struct {
	UserID       uint
	AccessToken  string
	RefreshToken string
	ATExpiresAt  time.Time
	RTExpiresAt  time.Time
}

// WorkOrder
type WorkOrder struct {
	ID          uint   `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
	Status      string `json:"status"`
	Unit        string `json:"unit"`
	PhotoURL    string `json:"photo"`

	RequesterID   uint   `json:"requesterId"`
	RequesterName string `json:"requester"`
	RequesterData User   `json:"requesterData"` // Akan diisi manual via JOIN

	AssigneeID *uint `json:"assigneeId"`
	Assignee   User  `json:"assignee"` // Akan diisi manual via JOIN

	// === TRACKING ===
	TakenAt       *time.Time `json:"taken_at"`
	CompletedAt   *time.Time `json:"completed_at"`
	CompletedByID *uint      `json:"completedById"`
	CompletedBy   User       `json:"completedBy"` // Akan diisi manual via JOIN

	CompletionNote string `json:"completion_note"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type DashboardStats struct {
	Incoming   int `json:"incoming"`
	Outgoing   int `json:"outgoing"`
	Pending    int `json:"pending"`
	InProgress int `json:"in_progress"`
}

// ActivityLog
type ActivityLog struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"userId"`
	UserName  string    `json:"userName"`
	Action    string    `json:"action"`
	RequestID uint      `json:"requestId"`
	Details   string    `json:"details"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
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
	PhotoURL    string `json:"photo"`
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

type FinalizeRequest struct {
	Note string `json:"note"`
}

type PaginationMeta struct {
	CurrentPage int `json:"current_page"`
	TotalPages  int `json:"total_pages"`
	TotalItems  int `json:"total_items"`
	Limit       int `json:"limit"`
}

type PaginatedResponse struct {
	Data interface{}    `json:"data"`
	Meta PaginationMeta `json:"meta"`
}
