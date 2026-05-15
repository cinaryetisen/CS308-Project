package models

import (
	"time"
)

// Refund represents a customer's refund request against a specific order item.
// The sales manager approves or rejects it, which is what flips Status and stamps ResolvedAt/ResolverID.
type Refund struct {
	ID           uint    `gorm:"primaryKey" json:"id"`
	OrderID      uint    `gorm:"not null" json:"order_id"`
	OrderItemID  uint    `gorm:"not null" json:"order_item_id"`
	CustomerID   uint    `gorm:"not null" json:"customer_id"`
	RefundAmount float64 `gorm:"not null" json:"refund_amount"` // Snapshot of OrderItem.Price so post-campaign discounts stay honored (req. 15)
	Reason       string  `gorm:"not null" json:"reason"`

	// Status can be: "pending", "approved", or "rejected"
	Status string `gorm:"default:'pending'" json:"status"`

	CreatedAt  time.Time  `json:"created_at"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
	ResolverID *uint      `json:"resolver_id,omitempty"` // Sales manager who decided
}
