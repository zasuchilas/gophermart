package models

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

type Order struct {
	OrderNum   string  `json:"number"`
	Status     string  `json:"status"`
	Accrual    float64 `json:"accrual"`
	UploadedAt string  `json:"uploaded_at"`
}

type UserBalance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type WithdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

type WithdrawalsData []*Withdrawal

type Withdrawal struct {
	OrderNum    string  `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

type OrderRow struct {
	ID         int64
	OrderNum   string
	Status     string
	Accrual    float64
	UserID     int64
	UploadedAt string
}

type OrderStateResponse struct {
	OrderNum string  `json:"number"`
	Status   string  `json:"status"`
	Accrual  float64 `json:"accrual"`
}
