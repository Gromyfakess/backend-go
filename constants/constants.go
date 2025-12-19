package constants

const (
	// Roles
	RoleAdmin = "Admin"
	RoleStaff = "Staff"

	// Status WorkOrder
	StatusPending    = "Pending"
	StatusInProgress = "In Progress"
	StatusCompleted  = "Completed"

	// Priority
	PriorityHigh   = "High"
	PriorityMedium = "Medium"
	PriorityLow    = "Low"

	// Availability
	AvailOnline  = "Online"
	AvailBusy    = "Busy"
	AvailAway    = "Away"
	AvailOffline = "Offline"

	// Default Config
	DefaultUnit = "IT Center"
	MaxFileSize = 2 << 20 // 2MB
)
