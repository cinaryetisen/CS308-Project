package controllers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"medieval-store/config"
	"medieval-store/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Function to get details of a specific product
func GetProduct(c *gin.Context) {
	//Get product ID from the URL
	idParam := c.Param("id")

	//Convert string ID into MongoDB ObjectID
	objectID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errro": "Invalid product ID format"})
		return
	}

	//Connect to MongoDB collection
	collection := config.MongoClient.Database("medieval_store").Collection("products")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//Query database for specific ID
	var product models.Product
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&product)

	//Handle errors
	if err != nil {
		//If MongoDB could not find the queried document
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}

		//If something else went wrong
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product"})
		return
	}

	//Return product to frontend
	c.JSON(http.StatusOK, product)
}

func GetProducts(c *gin.Context) {
	collection := config.MongoClient.Database("medieval_store").Collection("products")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Get query parameters from the URL
	searchQuery := c.Query("search")
	sortParam := strings.ToLower(c.Query("sort"))

	// 2. Build the MongoDB Filter (Search by Name OR Description)
	filter := bson.M{}

	if searchQuery != "" {
		regex := primitive.Regex{Pattern: searchQuery, Options: "i"}

		filter = bson.M{
			"$or": []bson.M{
				{"name": bson.M{"$regex": regex}},
				{"description": bson.M{"$regex": regex}},
			},
		}
	}

	// 3. Build the MongoDB Sort Options
	findOptions := options.Find()

	switch sortParam {
	case "price_asc":
		// Cheapest to Most Expensive
		findOptions.SetSort(bson.D{{Key: "price", Value: 1}})
	case "price_desc":
		// Most Expensive to Cheapest
		findOptions.SetSort(bson.D{{Key: "price", Value: -1}})
	case "popular":
		// Highest Review Count first
		findOptions.SetSort(bson.D{{Key: "review_count", Value: -1}})
	default:
		// Default sorting (e.g., newest items first)
		findOptions.SetSort(bson.D{{Key: "created_at", Value: -1}})
	}

	// 4. Execute the query WITH the filter and sort options
	cursor, err := collection.Find(ctx, filter, findOptions)
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

	// 5. Ensure we return an empty array instead of 'null' if nothing matches
	if products == nil {
		products = []models.Product{}
	}

	// Send the filtered and sorted data to the frontend
	c.JSON(http.StatusOK, products)
}
