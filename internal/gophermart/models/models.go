package models

import "time"

type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginData struct {
	UserID       int64
	Login        string
	PasswordHash string
}

type OrderData []*Order

type Order struct {
	Number     int       `json:"number"`
	Status     string    `json:"status"`
	Accrual    float32   `json:"accrual"`
	UploadedAt time.Time `json:"uploaded_at"`
}
