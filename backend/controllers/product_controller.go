package controllers

import (
	"context"
	"net/http"
	"time"

	"medieval-store/config" // Notice: The models import is gone!

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Define the Product structure directly here so Go knows what to expect
type Product struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Price       float64            `bson:"price" json:"price"`
	Description string             `bson:"description" json:"description"`
}

func GetProducts(c *gin.Context) {
	collection := config.GetCollection("products")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Grab everything from the database
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
		return
	}
	defer cursor.Close(ctx)

	// Use the locally defined Product struct
	var products []Product
	if err = cursor.All(ctx, &products); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode products"})
		return
	}

	// Send the data to the frontend!
	c.JSON(http.StatusOK, products)
}
