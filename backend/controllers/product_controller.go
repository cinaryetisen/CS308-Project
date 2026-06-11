package controllers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"medieval-store/config"
	"medieval-store/errs"
	"medieval-store/models"
	"medieval-store/services"

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
		errs.Abort(c, errs.ProductInvalidID)
		return
	}

	//Connect to MongoDB collection
	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//Query database for specific ID (soft-deleted products read as not found;
	//pending-price products too, unless a manager panel asks for them)
	detailFilter := bson.M{"_id": objectID, "deleted_at": nil}
	if c.Query("include_pending") != "true" {
		detailFilter["price_pending"] = bson.M{"$ne": true}
	}
	var product models.Product
	err = collection.FindOne(ctx, detailFilter).Decode(&product)

	//Handle errors
	if err != nil {
		if err == mongo.ErrNoDocuments {
			errs.Abort(c, errs.ProductNotFound)
			return
		}
		errs.Abort(c, errs.InternalError)
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
	// Soft-deleted products are excluded from every public listing.
	// {deleted_at: nil} matches both a missing field and an explicit null.
	andConditions := []bson.M{
		{"deleted_at": nil},
	}

	// Products awaiting a sales-manager price are hidden from the storefront.
	// Manager panels pass include_pending=true to keep seeing them.
	if c.Query("include_pending") != "true" {
		andConditions = append(andConditions, bson.M{"price_pending": bson.M{"$ne": true}})
	}

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
		errs.Abort(c, errs.InternalError)
		return
	}
	defer cursor.Close(ctx)

	// 4. Decode into your unmodified Go struct
	var products []models.Product
	if err = cursor.All(ctx, &products); err != nil {
		errs.Abort(c, errs.InternalError)
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
		errs.AbortWithDetail(c, errs.InvalidJSON, err.Error())
		return
	}

	objID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		errs.Abort(c, errs.ProductInvalidID)
		return
	}

	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	// Setting a real price also clears the pending flag, publishing the product.
	result, err := collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objID},
		bson.M{"$set": bson.M{"price": input.Price, "price_pending": false, "updated_at": time.Now()}},
	)
	if err != nil {
		errs.Abort(c, errs.InternalError)
		return
	}
	if result.MatchedCount == 0 {
		errs.Abort(c, errs.ProductNotFound)
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
		errs.AbortWithDetail(c, errs.InvalidJSON, err.Error())
		return
	}

	if *input.Discount < 0 || *input.Discount > 100 {
		errs.AbortWithDetail(c, errs.InvalidJSON, "discount must be between 0 and 100")
		return
	}

	objID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		errs.Abort(c, errs.ProductInvalidID)
		return
	}

	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")

	// Fetch current product to get name and price for the notification email.
	var product models.Product
	if err := collection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&product); err != nil {
		errs.Abort(c, errs.ProductNotFound)
		return
	}

	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objID},
		bson.M{"$set": bson.M{"discount": *input.Discount, "updated_at": time.Now()}},
	)
	if err != nil {
		errs.Abort(c, errs.InternalError)
		return
	}

	if *input.Discount > 0 {
		go services.NotifyWishlistOfDiscount(product, *input.Discount)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Discount updated successfully",
		"discount": *input.Discount,
	})
}
