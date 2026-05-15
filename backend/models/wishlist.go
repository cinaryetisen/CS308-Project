package models

import "time"

type WishlistItem struct {
	UserID    uint      `gorm:"primaryKey;autoIncrement:false" json:"user_id"`
	ProductID string    `gorm:"primaryKey;autoIncrement:false" json:"product_id"`
	CreatedAt time.Time `json:"created_at"`
}
