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
	"go.mongodb.org/mongo-driver/mongo"
)

// CreateRating submits or updates the calling user's rating for a delivered product.
// Invariant: at most one rating per (product_id, user_id) — second submission updates
// in place, recomputing the running average without changing review_count.
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

	var deliveredCount int64
	err := config.DB.Table("orders").
		Joins("JOIN order_items ON order_items.order_id = orders.id").
		Where("orders.customer_id = ? AND orders.status = ? AND order_items.product_id = ?", userID, "delivered", input.ProductID).
		Count(&deliveredCount).Error

	if err != nil || deliveredCount == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only rate products that have been delivered to you."})
		return
	}

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify user"})
		return
	}

	objID, _ := primitive.ObjectIDFromHex(input.ProductID)
	ratingsColl := config.MongoClient.Database(config.MongoDBName).Collection("ratings")
	productsColl := config.MongoClient.Database(config.MongoDBName).Collection("products")

	var existing models.Rating
	lookupErr := ratingsColl.FindOne(
		context.Background(),
		bson.M{"product_id": objID, "user_id": userID},
	).Decode(&existing)

	switch lookupErr {
	case nil:
		// Update path: this user already rated this product. Replace their rating
		// in place and re-roll the product's running average without changing review_count.
		_, updErr := ratingsColl.UpdateOne(
			context.Background(),
			bson.M{"_id": existing.ID},
			bson.M{"$set": bson.M{
				"rating":     input.Rating,
				"updated_at": time.Now(),
			}},
		)
		if updErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update rating"})
			return
		}

		var product models.Product
		if pErr := productsColl.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&product); pErr == nil && product.ReviewCount > 0 {
			newAvg := recomputeAvgOnUpdate(product.Rating, product.ReviewCount, existing.Rating, input.Rating)
			productsColl.UpdateOne(
				context.Background(),
				bson.M{"_id": objID},
				bson.M{"$set": bson.M{"rating": newAvg}},
			)
		}

		c.JSON(http.StatusOK, gin.H{"message": "Rating updated"})
		return

	case mongo.ErrNoDocuments:
		// Insert path: first time this user rates this product. Bump review_count.
		now := time.Now()
		newRating := models.Rating{
			ProductID: objID,
			UserID:    userID,
			UserName:  user.Name,
			Rating:    input.Rating,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if _, iErr := ratingsColl.InsertOne(context.Background(), newRating); iErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit rating"})
			return
		}

		var product models.Product
		if pErr := productsColl.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&product); pErr == nil {
			newCount := product.ReviewCount + 1
			newAvg := ((product.Rating * float64(product.ReviewCount)) + float64(input.Rating)) / float64(newCount)
			productsColl.UpdateOne(
				context.Background(),
				bson.M{"_id": objID},
				bson.M{"$set": bson.M{
					"rating":       newAvg,
					"review_count": newCount,
				}},
			)
		}

		c.JSON(http.StatusCreated, gin.H{"message": "Rating submitted successfully."})
		return

	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query existing rating"})
		return
	}
}

// recomputeAvgOnUpdate returns the new running average after one user's rating is replaced.
// Count stays the same; only the contribution from one user changes.
func recomputeAvgOnUpdate(oldAvg float64, count int, oldUserRating, newUserRating int) float64 {
	if count == 0 {
		return float64(newUserRating)
	}
	return (oldAvg*float64(count) - float64(oldUserRating) + float64(newUserRating)) / float64(count)
}

// GetProductRatings returns every rating for a product (public).
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

// GetMyRatingForProduct returns the calling user's rating for a specific product, or 404 if none.
// Used by the frontend to pre-fill the star picker on the product detail page.
func GetMyRatingForProduct(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	productID := c.Param("productId")
	objID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	coll := config.MongoClient.Database(config.MongoDBName).Collection("ratings")
	var existing models.Rating
	err = coll.FindOne(c.Request.Context(), bson.M{"product_id": objID, "user_id": userID}).Decode(&existing)

	if err == mongo.ErrNoDocuments {
		c.JSON(http.StatusNotFound, gin.H{"error": "no rating yet"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "lookup failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"rating": existing.Rating, "id": existing.ID.Hex()})
}
