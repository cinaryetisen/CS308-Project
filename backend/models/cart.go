package models

import "time"

type CartItem struct {
	UserID    uint      `gorm:"primaryKey;autoIncrement:false" json:"user_id"`
	ProductID string    `gorm:"primaryKey;autoIncrement:false;" json:"product_id"`
	Quantity  int       `gorm:"not null;default:1" json:"quantity"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CartItemResponse struct {
	ProductID string `json:"product_id"`
	Name      string `json:"name"`

	// Price is the EFFECTIVE unit price (discount already applied) — what the
	// customer pays. OriginalPrice/Discount let the UI render a strike-through.
	Price         float64 `json:"price"`
	OriginalPrice float64 `json:"original_price"`
	Discount      float64 `json:"discount"`

	Quantity int     `json:"quantity"`
	Subtotal float64 `json:"subtotal"`
	ImageURL string  `json:"image_url"`
	Stock    int     `json:"stock"`
	Category string  `json:"category"`
}
