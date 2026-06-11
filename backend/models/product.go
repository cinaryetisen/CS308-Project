package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Product struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name         string             `bson:"name" json:"name"`
	Model        string             `bson:"model" json:"model"`
	SerialNumber string             `bson:"serial_number" json:"serial_number"`
	Description  string             `bson:"description" json:"description"`
	Quantity     int                `bson:"quantity" json:"quantity"`

	// Pricing & Sales
	Cost     float64 `bson:"cost" json:"cost"`         // NEW: How much the store pays for it
	Price    float64 `bson:"price" json:"price"`       // How much the customer pays
	Discount float64 `bson:"discount" json:"discount"` // e.g., 10.0 for 10% off

	// True from PM creation until the sales manager sets a real price.
	// Pending products are hidden from the public storefront and unpurchasable.
	PricePending bool `bson:"price_pending,omitempty" json:"price_pending,omitempty"`

	// Logistics
	Warranty    string `bson:"warranty" json:"warranty"`
	Distributor string `bson:"distributor" json:"distributor"`

	// Frontend Display & Filtering
	Category string   `bson:"category" json:"category"`
	ImageURL string   `bson:"image_url" json:"image_url"`
	Tags     []string `bson:"tags" json:"tags"`

	// Aggregated Rating Data
	Rating      float64 `bson:"rating" json:"rating"`             // e.g., 4.5
	ReviewCount int     `bson:"review_count" json:"review_count"` // e.g., 12 reviews

	// Timestamps
	CreatedAt time.Time  `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time  `bson:"updated_at" json:"updated_at"`
	DeletedAt *time.Time `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"` // NEW: For soft deletes`
}

// EffectivePrice is the price the customer actually pays: base price with the
// active discount applied. Every money-path (cart, checkout, invoices) must
// charge this — never raw Price — so the advertised discount is honored and
// the OrderItem.Price snapshot used for refunds stays correct.
func (p Product) EffectivePrice() float64 {
	return p.Price * (1 - p.Discount/100)
}
