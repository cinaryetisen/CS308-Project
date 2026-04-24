package controllers

import (
	"context"
	"net/http"
	"time"

	"medieval-store/config"
	"medieval-store/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Function to submit a review to the database for a delivered product
func CreateReview(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var input struct {
		ProductID string `json:"product_id" binding:"required"`
		Rating    int    `json:"rating" binding:"required,min=1,max=5"`
		Comment   string `json:"comment" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var count int64
	err := config.DB.Table("orders").
		Joins("JOIN order_items ON order_items.order_id = orders.id").
		Where("orders.customer_id = ? AND orders.status = ? AND order_items.product_id = ?", userID, "delivered", input.ProductID).
		Count(&count).Error

	if err != nil || count == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only review products that have been delivered to you."})
		return
	}

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify user"})
		return
	}

	objID, _ := primitive.ObjectIDFromHex(input.ProductID)
	review := models.Review{
		ProductID: objID,
		UserID:    userID,
		UserName:  user.Name,
		Rating:    input.Rating,
		Comment:   input.Comment,
		Status:    "pending", // Product Manager must approve this!
		CreatedAt: time.Now(),
	}

	collection := config.MongoClient.Database("medieval_store").Collection("reviews")
	_, err = collection.InsertOne(context.Background(), review)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit review"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Review submitted successfully and is awaiting moderation."})
}

// Function to get approved reviews for a product
func GetProductReviews(c *gin.Context) {
	productID := c.Param("product_id")
	objID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	collection := config.MongoClient.Database("medieval_store").Collection("reviews")

	filter := bson.M{"product_id": objID, "status": "approved"}
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reviews"})
		return
	}

	var reviews []models.Review
	if err = cursor.All(context.Background(), &reviews); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse reviews"})
		return
	}

	if reviews == nil {
		reviews = []models.Review{}
	}

	c.JSON(http.StatusOK, reviews)
}

// Function to get pending reviews for a product manager
func GetPendingReviews(c *gin.Context) {
	collection := config.MongoClient.Database("medieval_store").Collection("reviews")

	cursor, err := collection.Find(context.Background(), bson.M{"status": "pending"})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch pending reviews"})
		return
	}

	var reviews []models.Review
	if err = cursor.All(context.Background(), &reviews); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse reviews"})
		return
	}

	if reviews == nil {
		reviews = []models.Review{}
	}

	c.JSON(http.StatusOK, reviews)
}

// Function to change status of a review
func ModerateReview(c *gin.Context) {
	reviewID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(reviewID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID format"})
		return
	}

	var input struct {
		Action string `json:"action" binding:"required"` // "approve" or "reject"
	}
	if err := c.ShouldBindJSON(&input); err != nil || (input.Action != "approve" && input.Action != "reject") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Action must be 'approve' or 'reject'"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	reviewsCollection := config.MongoClient.Database("medieval_store").Collection("reviews")

	newStatus := "rejected"
	if input.Action == "approve" {
		newStatus = "approved"
	}

	var updatedReview models.Review
	err = reviewsCollection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": objID},
		bson.M{"$set": bson.M{"status": newStatus}},
	).Decode(&updatedReview)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Review not found or already processed"})
		return
	}

	if newStatus == "approved" {
		productsCollection := config.MongoClient.Database("medieval_store").Collection("products")

		var product models.Product
		if err := productsCollection.FindOne(ctx, bson.M{"_id": updatedReview.ProductID}).Decode(&product); err == nil {
			newCount := product.ReviewCount + 1
			newRating := ((product.Rating * float64(product.ReviewCount)) + float64(updatedReview.Rating)) / float64(newCount)

			productsCollection.UpdateOne(
				ctx,
				bson.M{"_id": updatedReview.ProductID},
				bson.M{"$set": bson.M{
					"rating":       newRating,
					"review_count": newCount,
				}},
			)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Review " + newStatus + " successfully"})
}
