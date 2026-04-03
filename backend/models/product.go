package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Product struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name         string             `bson:"name" json:"name"`
	Model        string             `bson:"model" json:"model"`
	SerialNumber string             `bson:"serial_number" json:"serial_number"`
	Description  string             `bson:"description" json:"description"`
	Quantity     int                `bson:"quantity" json:"quantity"`

	// Pricing & Sales
	Price    float64 `bson:"price" json:"price"`
	Discount float64 `bson:"discount" json:"discount"` // e.g., 10.0 for 10% off

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
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}
