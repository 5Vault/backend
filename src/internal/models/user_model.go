package models

type RequestUser struct {
	Username string `json:"username" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Phone    string `json:"phone" binding:"required"`
}

type ResponseUser struct {
	UserID        string  `json:"user_id"`
	Username      string  `json:"username"`
	Name          string  `json:"name"`
	Email         string  `json:"email"`
	Phone         string  `json:"phone"`
	ApiKey        *string `json:"api_key,omitempty"`
	Tier          string  `json:"tier,omitempty"`
	TierName      string  `json:"tier_name,omitempty"`
	TierUpdatedAt string  `json:"tier_update_at,omitempty"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at,omitempty"`
	DeletedAt     string  `json:"deleted_at,omitempty"`
}

type UserDashboard struct {
	TotalFiles   int64  `json:"total_files"`
	TotalStorage int64  `json:"total_storage"`
	TotalSize    int64  `json:"total_size"`
	RecentFiles  []File `json:"recent_files"`
}
