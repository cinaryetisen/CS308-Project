package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Review struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProductID primitive.ObjectID `bson:"product_id" json:"product_id"` // Links to MongoDB Product

	// User Info (Linking to PostgreSQL)
	UserID   uint   `bson:"user_id" json:"user_id"`     // Links to PostgreSQL User ID
	UserName string `bson:"user_name" json:"user_name"` // "Denormalized" so we don't have to query Postgres just to show the name

	// The Content
	Rating  int    `bson:"rating" json:"rating"` // e.g., 1 through 5
	Comment string `bson:"comment" json:"comment"`

	CreatedAt time.Time `bson:"created_at" json:"created_at"`
}
