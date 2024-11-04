package models

type RegisterGoodsRequest struct {
	Match      string `json:"match"`
	Reward     int    `json:"reward"`
	RewardType string `json:"reward_type"`
}

type RegisterOrderRequest struct {
	Order string `json:"order"`
	Goods []GoodsPosition
}

type GoodsPosition struct {
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}
