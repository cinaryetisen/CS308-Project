package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Product struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name         string             `bson:"name" json:"name"`
	Model        string             `bson:"model" json:"model"`
	SerialNumber string             `bson:"serial_number" json:"serial_number"`
	Description  string             `bson:"description" json:"description"`
	Quantity     int                `bson:"quantity" json:"quantity"`
	Price        float64            `bson:"price" json:"price"`
	Warranty     string             `bson:"warranty" json:"warranty"`
	Distributor  string             `bson:"distributor" json:"distributor"`
	Category     string             `bson:"category" json:"category"`
	ImageURL     string             `bson:"image_url" json:"image_url"`
	Tags         []string           `bson:"tags" json:"tags"`
}