package models

import "time"

type CartItem struct {
	UserID    uint      `gorm:"primaryKey;autoIncrement:false" json:"user_id"`
	ProductID string    `gorm:"primaryKey;autoIncrement:false;" json:"product_id"`
	Quantity  int       `gorm:"not null;default:1" json:"quantity"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CartItemResponse struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
	Subtotal  float64 `json:"subtotal"`
}
