package controllers

import (
	"context"
	"net/http"
	"time"

	"medieval-store/config"
	"medieval-store/errs"
	"medieval-store/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// B1: Create a new product (Product Manager Only)
func CreateProduct(c *gin.Context) {
	// 1. Strict Input DTO: Blocks users from injecting fake ratings or discounts!
	var input struct {
		Name         string   `json:"name" binding:"required"`
		Model        string   `json:"model" binding:"required"`
		SerialNumber string   `json:"serial_number" binding:"required"`
		Description  string   `json:"description" binding:"required"`
		Quantity     int      `json:"quantity" binding:"gte=0"`      // Must be 0 or greater
		Price        float64  `json:"price" binding:"required,gt=0"` // Must be greater than 0
		Category     string   `json:"category" binding:"required"`
		Distributor  string   `json:"distributor" binding:"required"`
		Warranty     string   `json:"warranty" binding:"required"`
		ImageURL     string   `json:"image_url"`
		Tags         []string `json:"tags"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		errs.AbortWithDetail(c, errs.InvalidJSON, err.Error())
		return
	}

	// 2. Map the safe input to your actual MongoDB model
	product := models.Product{
		ID:           primitive.NewObjectID(),
		Name:         input.Name,
		Model:        input.Model,
		SerialNumber: input.SerialNumber,
		Description:  input.Description,
		Quantity:     input.Quantity,
		Price:        input.Price,
		Cost:         input.Price * 0.6, // Auto-calculate cost for the profit algorithm!
		Discount:     0.0,               // Force starting discount to 0
		Category:     input.Category,
		Distributor:  input.Distributor,
		Warranty:     input.Warranty,
		ImageURL:     input.ImageURL,
		Tags:         input.Tags,
		Rating:       0.0, // Force rating to 0
		ReviewCount:  0,   // Force reviews to 0
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	collection := config.MongoClient.Database("medieval_store").Collection("products")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, product)
	if err != nil {
		errs.Abort(c, errs.InternalError)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Product created successfully", "product": product})
}

// B2: Update non-price fields (Product Manager Only)
func UpdateProduct(c *gin.Context) {
	idParam := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		errs.Abort(c, errs.ProductInvalidID)
		return
	}

	// Explicit allowlist: Price, Cost, Discount, and Quantity are completely locked out here!
	var input struct {
		Name         string   `json:"name"`
		Model        string   `json:"model"`
		SerialNumber string   `json:"serial_number"`
		Description  string   `json:"description"`
		Category     string   `json:"category"`
		Distributor  string   `json:"distributor"`
		Warranty     string   `json:"warranty"`
		ImageURL     string   `json:"image_url"`
		Tags         []string `json:"tags"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		errs.Abort(c, errs.InvalidJSON)
		return
	}

	updateData := bson.M{
		"$set": bson.M{
			"name":          input.Name,
			"model":         input.Model,
			"serial_number": input.SerialNumber,
			"description":   input.Description,
			"category":      input.Category,
			"distributor":   input.Distributor,
			"warranty":      input.Warranty,
			"image_url":     input.ImageURL,
			"tags":          input.Tags,
			"updated_at":    time.Now(),
		},
	}

	collection := config.MongoClient.Database("medieval_store").Collection("products")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Only update if it hasn't been soft-deleted
	filter := bson.M{"_id": objectID, "deleted_at": bson.M{"$exists": false}}
	result, err := collection.UpdateOne(ctx, filter, updateData)

	if err != nil {
		errs.Abort(c, errs.InternalError)
		return
	}
	if result.MatchedCount == 0 {
		errs.Abort(c, errs.ProductNotFound)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product details updated successfully"})
}

// B3: Soft Delete a product (Product Manager Only)
func DeleteProduct(c *gin.Context) {
	idParam := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		errs.Abort(c, errs.ProductInvalidID)
		return
	}

	collection := config.MongoClient.Database("medieval_store").Collection("products")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Perform a soft delete by setting the DeletedAt timestamp (Matches your model exactly)
	updateData := bson.M{"$set": bson.M{"deleted_at": time.Now()}}
	result, err := collection.UpdateOne(ctx, bson.M{"_id": objectID}, updateData)

	if err != nil {
		errs.Abort(c, errs.InternalError)
		return
	}
	if result.MatchedCount == 0 {
		errs.Abort(c, errs.ProductNotFound)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product successfully removed from catalog"})
}

// B4: Adjust Stock Levels safely (Product Manager Only)
func UpdateStock(c *gin.Context) {
	idParam := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		errs.Abort(c, errs.ProductInvalidID)
		return
	}

	// Accept a positive or negative delta (e.g., {"delta": 5} or {"delta": -2})
	var input struct {
		Delta int `json:"delta" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		errs.AbortWithDetail(c, errs.InvalidJSON, "please provide a valid stock delta")
		return
	}

	collection := config.MongoClient.Database("medieval_store").Collection("products")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": objectID, "deleted_at": bson.M{"$exists": false}}

	// If decreasing stock, guarantee the database has enough stock to fulfill the request
	if input.Delta < 0 {
		filter["quantity"] = bson.M{"$gte": -input.Delta}
	}

	// Atomically increment or decrement the stock, and update the timestamp
	update := bson.M{"$inc": bson.M{"quantity": input.Delta}, "$set": bson.M{"updated_at": time.Now()}}
	result, err := collection.UpdateOne(ctx, filter, update)

	if err != nil {
		errs.Abort(c, errs.InternalError)
		return
	}

	if result.MatchedCount == 0 {
		errs.Abort(c, errs.ProductOutOfStock)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Stock updated successfully"})
}
