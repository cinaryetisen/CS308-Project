package controllers

import (
	"context"
	"net/http"
	"sort"
	"strings"
	"time"

	"medieval-store/config"
	"medieval-store/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func GetProducts(c *gin.Context) {
	collection := config.GetCollection("products")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
		return
	}
	defer cursor.Close(ctx)

	var products []models.Product

	if err = cursor.All(ctx, &products); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode products"})
		return
	}

	// Sorting algm. for the prices
	sortParam := strings.ToLower(c.Query("sort"))

	if sortParam == "asc" {
		// Cheapest to Expensive
		sort.Slice(products, func(i, j int) bool {
			return products[i].Price < products[j].Price
		})
	} else if sortParam == "desc" {
		//Expensive to Cheapest
		sort.Slice(products, func(i, j int) bool {
			return products[i].Price > products[j].Price
		})
	}
	c.JSON(http.StatusOK, products)
}
