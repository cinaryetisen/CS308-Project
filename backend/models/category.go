package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Category is the canonical list of product categories. Product.Category is kept
// as a string (no FK) for minimal migration churn; the validation that a product's
// Category exists in this collection happens in the admin product handlers (B1/B2).
type Category struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}
