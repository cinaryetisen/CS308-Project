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
	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
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
	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Get query parameters from the URL
	searchQuery := c.Query("search")
	sortParam := strings.ToLower(c.Query("sort"))
	categoryParam := c.Query("category")

	// 2. Set up the Aggregation Pipeline
	pipeline := mongo.Pipeline{}

	// Stage A: Search & Category Filter ($match)
	var andConditions []bson.M

	if searchQuery != "" {
		regex := primitive.Regex{Pattern: searchQuery, Options: "i"}
		andConditions = append(andConditions, bson.M{
			"$or": []bson.M{
				{"name": bson.M{"$regex": regex}},
				{"description": bson.M{"$regex": regex}},
			},
		})
	}

	if categoryParam != "" && categoryParam != "All" {
		andConditions = append(andConditions, bson.M{"category": categoryParam})
	}

	filter := bson.M{}
	if len(andConditions) > 0 {
		filter["$and"] = andConditions
	}

	pipeline = append(pipeline, bson.D{{Key: "$match", Value: filter}})

	// Stage B: Calculate TEMPORARY variables for sorting ($addFields)
	pipeline = append(pipeline, bson.D{{
		Key: "$addFields",
		Value: bson.M{
			// tmp_sort_price = price - (price * (discount / 100))
			"tmp_sort_price": bson.M{
				"$subtract": bson.A{
					"$price",
					bson.M{"$multiply": bson.A{
						"$price",
						bson.M{"$divide": bson.A{"$discount", 100}},
					}},
				},
			},
			// popularity_score = rating * review_count
			"popularity_score": bson.M{
				"$multiply": bson.A{"$rating", "$review_count"},
			},
		},
	}})

	// Stage C: Sort using the temporary variables ($sort)
	var sortStage bson.D
	switch sortParam {
	case "price_asc":
		sortStage = bson.D{{Key: "tmp_sort_price", Value: 1}}
	case "price_desc":
		sortStage = bson.D{{Key: "tmp_sort_price", Value: -1}}
	case "popular":
		sortStage = bson.D{{Key: "popularity_score", Value: -1}}
	default:
		sortStage = bson.D{{Key: "created_at", Value: -1}} // Default sorting
	}
	pipeline = append(pipeline, bson.D{{Key: "$sort", Value: sortStage}})

	// Stage D: Destroy the temporary variables ($unset)
	// This ensures the data perfectly matches your existing Go Product struct!
	pipeline = append(pipeline, bson.D{{Key: "$unset", Value: bson.A{"tmp_sort_price", "popularity_score"}}})

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

// Update a product's base price (sales manager only).
func UpdateProductPrice(c *gin.Context) {
	productID := c.Param("id")

	var input struct {
		Price float64 `json:"price" binding:"required,gt=0"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	objID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	result, err := collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objID},
		bson.M{"$set": bson.M{"price": input.Price, "updated_at": time.Now()}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update price"})
		return
	}
	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Price updated successfully", "price": input.Price})
}

// Set a discount percentage on a product and notifie wishlist users.
func SetProductDiscount(c *gin.Context) {
	productID := c.Param("id")

	// Use a pointer so 0.0 (removing a discount) is distinguishable from "not provided".
	var input struct {
		Discount *float64 `json:"discount" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if *input.Discount < 0 || *input.Discount > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Discount must be between 0 and 100"})
		return
	}

	objID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")

	// Fetch current product to get name and price for the notification email.
	var product models.Product
	if err := collection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&product); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objID},
		bson.M{"$set": bson.M{"discount": *input.Discount, "updated_at": time.Now()}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update discount"})
		return
	}

	if *input.Discount > 0 {
		// TODO: Notify wishlist users asynchronously when a real discount is applied.
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Discount updated successfully",
		"discount": *input.Discount,
	})
}
