package models

import (
	"time"
)

// User represents a customer or manager in the store
type User struct {
	ID          uint      `gorm:"primaryKey"`
	Name        string    `gorm:"not null" json:"name"`
	TaxID       string    `json:"tax_id"`
	Email       string    `gorm:"unique;not null" json:"email"`
	HomeAddress string    `json:"home_address"`
	Password    string    `gorm:"not null" json:"password"` // Stores the bcrypt hash
	Role        string    `gorm:"default:'customer'" json:"role"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
