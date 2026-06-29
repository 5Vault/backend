package models

type RequestUser struct {
	Username string `json:"username" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Phone    string `json:"phone"`
}

type ResponseUser struct {
	UserID               string  `json:"user_id"`
	Username             string  `json:"username"`
	Name                 string  `json:"name"`
	Email                string  `json:"email"`
	Phone                string  `json:"phone,omitempty"`
	ApiKey               *string `json:"api_key,omitempty"`
	Tier                 string  `json:"tier,omitempty"`
	TierName             string  `json:"tier_name,omitempty"`
	TierUpdatedAt        string  `json:"tier_updated_at,omitempty"`
	ExtraStorageEnabled  bool    `json:"extra_storage_enabled"`
	TwoFAEnabled         bool    `json:"two_fa_enabled"`
	AvatarURL            *string `json:"avatar_url,omitempty"`
	CreatedAt            string  `json:"created_at"`
	UpdatedAt            string  `json:"updated_at,omitempty"`
}

type UserDashboard struct {
	TotalFiles   int64  `json:"total_files"`
	TotalStorage int64  `json:"total_storage"`
	TotalSize    int64  `json:"total_size"`
	RecentFiles  []File `json:"recent_files"`
}
