package models

type RequestLogin struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

type ResponseToken struct {
	UserID string `json:"user_id"`
	Token  string `json:"token"`
}
