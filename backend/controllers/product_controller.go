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
)

// Function to get details of a specific product
func GetProduct(c *gin.Context) {
	//Get product ID from the URL
	idParam := c.Param("id")

	//Convert string ID into MongoDB ObjectID
	objectID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID format"})
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

	// 2. Set up the Aggregation Pipeline
	pipeline := mongo.Pipeline{}

	// Stage A: Search Filter ($match)
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
	pipeline = append(pipeline, bson.D{{Key: "$match", Value: filter}})

	// Stage B: Calculate a TEMPORARY price for sorting only ($addFields)
	// tmp_sort_price = price - (price * (discount / 100))
	pipeline = append(pipeline, bson.D{{
		Key: "$addFields",
		Value: bson.M{
			"tmp_sort_price": bson.M{
				"$subtract": bson.A{
					"$price",
					bson.M{"$multiply": bson.A{
						"$price",
						bson.M{"$divide": bson.A{"$discount", 100}},
					}},
				},
			},
		},
	}})

	// Stage C: Sort using the temporary variable ($sort)
	var sortStage bson.D
	switch sortParam {
	case "price_asc":
		sortStage = bson.D{{Key: "tmp_sort_price", Value: 1}}
	case "price_desc":
		sortStage = bson.D{{Key: "tmp_sort_price", Value: -1}}
	case "popular":
		sortStage = bson.D{{Key: "review_count", Value: -1}}
	default:
		sortStage = bson.D{{Key: "created_at", Value: -1}} // Default sorting
	}
	pipeline = append(pipeline, bson.D{{Key: "$sort", Value: sortStage}})

	// Stage D: Destroy the temporary variable ($unset)
	// This ensures the data perfectly matches your existing Go Product struct!
	pipeline = append(pipeline, bson.D{{Key: "$unset", Value: "tmp_sort_price"}})

	// 3. Execute the Pipeline
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
		return
	}
	defer cursor.Close(ctx)

	// 4. Decode into your unmodified Go struct
	var products []models.Product
	if err = cursor.All(ctx, &products); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode products"})
		return
	}

	if products == nil {
		products = []models.Product{}
	}

	// Send the correctly sorted data to the frontend
	c.JSON(http.StatusOK, products)
}
