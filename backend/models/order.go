package models

import (
	"time"

	"medieval-store/security"

	"gorm.io/gorm"
)

// Order represents the overall delivery and invoice
type Order struct {
	ID              uint    `gorm:"primaryKey" json:"delivery_id"` // Serves as the delivery ID
	CustomerID      uint    `gorm:"not null" json:"customer_id"`
	TotalPrice      float64 `gorm:"not null" json:"total_price"`
	DeliveryAddress string  `gorm:"not null" json:"delivery_address"`

	// Status can be: "processing", "in-transit", "delivered", "cancelled", or "returned"
	Status    string `gorm:"default:'processing'" json:"status"`
	Completed bool   `gorm:"default:false" json:"completed"`

	CreatedAt time.Time `json:"created_at"` // Crucial for the 30-day refund window
	UpdatedAt time.Time `json:"updated_at"`

	// Relational setup: One Order has many OrderItems
	Items []OrderItem `gorm:"foreignKey:OrderID" json:"items"`
}

// OrderItem represents the individual products within a specific delivery
type OrderItem struct {
	ID        uint    `gorm:"primaryKey" json:"id"`
	OrderID   uint    `gorm:"not null" json:"order_id"`
	ProductID string  `gorm:"not null" json:"product_id"` // Change to uint if your Product ID is an integer
	Quantity  int     `gorm:"not null" json:"quantity"`
	Price     float64 `gorm:"not null" json:"price"` // Price at the time of purchase (important for discounts)
}

// BeforeSave encrypts the delivery address before it hits the DB column.
func (o *Order) BeforeSave(tx *gorm.DB) error {
	enc, err := security.Encrypt(o.DeliveryAddress)
	if err != nil {
		return err
	}
	o.DeliveryAddress = enc
	return nil
}

// AfterSave decrypts the address back into the struct so callers keep their plaintext copy.
func (o *Order) AfterSave(tx *gorm.DB) error {
	dec, err := security.Decrypt(o.DeliveryAddress)
	if err != nil {
		return err
	}
	o.DeliveryAddress = dec
	return nil
}

// AfterFind decrypts on read so handlers never see ciphertext.
func (o *Order) AfterFind(tx *gorm.DB) error {
	dec, err := security.Decrypt(o.DeliveryAddress)
	if err != nil {
		return err
	}
	o.DeliveryAddress = dec
	return nil
}
