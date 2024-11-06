package models

import (
	"github.com/Rhymond/go-money"
	"time"
)

type GoodsData struct {
	Match      string  `json:"match"`
	Reward     float64 `json:"reward"`
	RewardType string  `json:"reward_type"`
}

type Receipt struct {
	Order string `json:"order"`
	Goods []GoodsPosition
}

type GoodsPosition struct {
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

type OrderData struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

type AccrualOrder struct {
	ID         int64
	OrderNum   string
	Status     string
	Accrual    *money.Money
	Receipt    *Receipt
	UploadedAt time.Time
}
