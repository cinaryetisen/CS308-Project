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

// categoryExists reports whether a category with this exact name is registered.
// Product create/update validate against this so the catalog can't reference
// categories that were never created (or were deleted) in the admin panel.
func categoryExists(name string) (bool, error) {
	collection := config.MongoClient.Database(config.MongoDBName).Collection("categories")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	count, err := collection.CountDocuments(ctx, bson.M{"name": name})
	return count > 0, err
}

// B1: Create a new product (Product Manager Only)
func CreateProduct(c *gin.Context) {
	var input struct {
		Name         string   `json:"name" binding:"required"`
		Model        string   `json:"model" binding:"required"`
		SerialNumber string   `json:"serial_number" binding:"required"`
		Description  string   `json:"description" binding:"required"`
		Quantity     int      `json:"quantity" binding:"gte=0"`
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

	// The category must be one registered through the admin category endpoints.
	exists, err := categoryExists(input.Category)
	if err != nil {
		errs.Abort(c, errs.InternalError)
		return
	}
	if !exists {
		errs.Abort(c, errs.CategoryNotFound)
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
		Price:        99999.99,
		PricePending: true, // hidden from the storefront until Sales sets a real price
		Cost:         0.0,
		Discount:     0.0,
		Category:     input.Category,
		Distributor:  input.Distributor,
		Warranty:     input.Warranty,
		ImageURL:     input.ImageURL,
		Tags:         input.Tags,
		Rating:       0.0,
		ReviewCount:  0,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := collection.InsertOne(ctx, product); err != nil {
		errs.Abort(c, errs.InternalError)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Product created successfully! Price temporarily set to 99,999.99 until approved by Sales.",
		"product": product,
	})
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
	// Pointer fields give true PATCH semantics — only fields present in the JSON
	// are written, so a partial update can never blank out the omitted ones.
	var input struct {
		Name         *string   `json:"name"`
		Model        *string   `json:"model"`
		SerialNumber *string   `json:"serial_number"`
		Description  *string   `json:"description"`
		Category     *string   `json:"category"`
		Distributor  *string   `json:"distributor"`
		Warranty     *string   `json:"warranty"`
		ImageURL     *string   `json:"image_url"`
		Tags         *[]string `json:"tags"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		errs.Abort(c, errs.InvalidJSON)
		return
	}

	set := bson.M{"updated_at": time.Now()}
	if input.Name != nil {
		set["name"] = *input.Name
	}
	if input.Model != nil {
		set["model"] = *input.Model
	}
	if input.SerialNumber != nil {
		set["serial_number"] = *input.SerialNumber
	}
	if input.Description != nil {
		set["description"] = *input.Description
	}
	if input.Category != nil {
		exists, err := categoryExists(*input.Category)
		if err != nil {
			errs.Abort(c, errs.InternalError)
			return
		}
		if !exists {
			errs.Abort(c, errs.CategoryNotFound)
			return
		}
		set["category"] = *input.Category
	}
	if input.Distributor != nil {
		set["distributor"] = *input.Distributor
	}
	if input.Warranty != nil {
		set["warranty"] = *input.Warranty
	}
	if input.ImageURL != nil {
		set["image_url"] = *input.ImageURL
	}
	if input.Tags != nil {
		set["tags"] = *input.Tags
	}

	if len(set) == 1 { // only updated_at — nothing to change
		errs.AbortWithDetail(c, errs.InvalidJSON, "no updatable fields provided")
		return
	}

	updateData := bson.M{"$set": set}

	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
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

	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
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

	// Accept a positive or negative delta (e.g., {"delta": 5} or {"delta": -2}).
	// Pointer binding so a literal 0 is "present but invalid" (clear 400), not
	// confused with "field missing" by Go's zero-value handling.
	var input struct {
		Delta *int `json:"delta" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		errs.AbortWithDetail(c, errs.InvalidJSON, "please provide a valid stock delta")
		return
	}
	if *input.Delta == 0 {
		errs.AbortWithDetail(c, errs.InvalidJSON, "delta must be non-zero")
		return
	}
	delta := *input.Delta

	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": objectID, "deleted_at": bson.M{"$exists": false}}

	// If decreasing stock, guarantee the database has enough stock to fulfill the request
	if delta < 0 {
		filter["quantity"] = bson.M{"$gte": -delta}
	}

	// Atomically increment or decrement the stock, and update the timestamp
	update := bson.M{"$inc": bson.M{"quantity": delta}, "$set": bson.M{"updated_at": time.Now()}}
	result, err := collection.UpdateOne(ctx, filter, update)

	if err != nil {
		errs.Abort(c, errs.InternalError)
		return
	}

	if result.MatchedCount == 0 {
		// Distinguish "no such product" from "not enough stock" so the PM panel
		// can show an accurate error.
		count, cErr := collection.CountDocuments(ctx, bson.M{"_id": objectID, "deleted_at": bson.M{"$exists": false}})
		if cErr == nil && count == 0 {
			errs.Abort(c, errs.ProductNotFound)
			return
		}
		errs.Abort(c, errs.ProductOutOfStock)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Stock updated successfully"})
}
