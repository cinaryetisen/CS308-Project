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

// Function to submit a rating to the database for a delivered product
func CreateRating(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var input struct {
		ProductID string `json:"product_id" binding:"required"`
		Rating    int    `json:"rating" binding:"required,min=1,max=5"`
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
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only rate products that have been delivered to you."})
		return
	}

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify user"})
		return
	}

	objID, _ := primitive.ObjectIDFromHex(input.ProductID)
	rating := models.Rating{
		ProductID: objID,
		UserID:    userID,
		UserName:  user.Name,
		Rating:    input.Rating,
		CreatedAt: time.Now(),
	}

	collection := config.MongoClient.Database(config.MongoDBName).Collection("ratings")
	_, err = collection.InsertOne(context.Background(), rating)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit rating"})
		return
	}

	productsCollection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	var product models.Product
	if err := productsCollection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&product); err == nil {
		newCount := product.ReviewCount + 1
		newRating := ((product.Rating * float64(product.ReviewCount)) + float64(input.Rating)) / float64(newCount)

		productsCollection.UpdateOne(
			context.Background(),
			bson.M{"_id": objID},
			bson.M{"$set": bson.M{
				"rating":       newRating,
				"review_count": newCount,
			}},
		)
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Rating submitted successfully."})
}

// Function to get ratings for a product
func GetProductRatings(c *gin.Context) {
	productID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	collection := config.MongoClient.Database(config.MongoDBName).Collection("ratings")

	filter := bson.M{"product_id": objID}
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch ratings"})
		return
	}

	var ratings []models.Rating
	if err = cursor.All(context.Background(), &ratings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse ratings"})
		return
	}

	if ratings == nil {
		ratings = []models.Rating{}
	}

	c.JSON(http.StatusOK, ratings)
}
