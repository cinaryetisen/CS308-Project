package models

import "time"

type WishlistItem struct {
	UserID    uint      `gorm:"primaryKey;autoIncrement:false" json:"user_id"`
	ProductID string    `gorm:"primaryKey;autoIncrement:false" json:"product_id"`
	CreatedAt time.Time `json:"created_at"`
}

type WishlistItemResponse struct {
	ProductID string    `json:"product_id"`
	Name      string    `json:"name"`
	Price     float64   `json:"price"`
	Discount  float64   `json:"discount"`
	ImageURL  string    `json:"image_url"`
	Stock     int       `json:"stock"`
	Category  string    `json:"category"`
	AddedAt   time.Time `json:"added_at"`
}
