package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"medieval-store/config"
	"medieval-store/errs"
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
		errs.AbortWithDetail(c, errs.InvalidJSON, err.Error())
		return
	}

	var deliveredCount int64
	err := config.DB.Table("orders").
		Joins("JOIN order_items ON order_items.order_id = orders.id").
		Where("orders.customer_id = ? AND orders.status = ? AND order_items.product_id = ?", userID, "delivered", input.ProductID).
		Count(&deliveredCount).Error

	if err != nil || deliveredCount == 0 {
		errs.Abort(c, errs.RatingUnauthorized)
		return
	}

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		errs.Abort(c, errs.InternalError)
		return
	}

	objID, err := primitive.ObjectIDFromHex(input.ProductID)
	if err != nil {
		errs.Abort(c, errs.ProductInvalidID)
		return
	}

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
			errs.Abort(c, errs.InternalError)
			return
		}

		// Shift the running average by this user's delta in ONE atomic pipeline
		// update — the math runs inside MongoDB against the current document, so
		// concurrent raters can't interleave a stale read-modify-write.
		delta := float64(input.Rating - existing.Rating)
		safeCount := bson.M{"$ifNull": bson.A{"$review_count", 0}}
		avgShift := mongo.Pipeline{
			{{Key: "$set", Value: bson.M{
				"rating": bson.M{"$cond": bson.A{
					bson.M{"$gt": bson.A{safeCount, 0}},
					bson.M{"$add": bson.A{
						bson.M{"$ifNull": bson.A{"$rating", 0}},
						bson.M{"$divide": bson.A{delta, safeCount}},
					}},
					bson.M{"$ifNull": bson.A{"$rating", 0}},
				}},
			}}},
		}
		if _, err := productsColl.UpdateOne(context.Background(), bson.M{"_id": objID}, avgShift); err != nil {
			log.Printf("failed to update product average rating for %s: %v", objID.Hex(), err)
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
			errs.Abort(c, errs.InternalError)
			return
		}

		// Fold the new rating into the average in ONE atomic pipeline update.
		// All field references on the right-hand side read the pre-update values,
		// so: rating = (old_avg*old_count + new) / (old_count+1), count = old+1.
		safeRating := bson.M{"$ifNull": bson.A{"$rating", 0}}
		safeCount := bson.M{"$ifNull": bson.A{"$review_count", 0}}
		avgFold := mongo.Pipeline{
			{{Key: "$set", Value: bson.M{
				"rating": bson.M{"$divide": bson.A{
					bson.M{"$add": bson.A{
						bson.M{"$multiply": bson.A{safeRating, safeCount}},
						float64(input.Rating),
					}},
					bson.M{"$add": bson.A{safeCount, 1}},
				}},
				"review_count": bson.M{"$add": bson.A{safeCount, 1}},
			}}},
		}
		if _, err := productsColl.UpdateOne(context.Background(), bson.M{"_id": objID}, avgFold); err != nil {
			log.Printf("failed to update product average rating for %s: %v", objID.Hex(), err)
		}

		c.JSON(http.StatusCreated, gin.H{"message": "Rating submitted successfully."})
		return

	default:
		errs.Abort(c, errs.InternalError)
		return
	}
}

// GetProductRatings returns every rating for a product (public).
func GetProductRatings(c *gin.Context) {
	productID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		errs.Abort(c, errs.ProductInvalidID)
		return
	}

	collection := config.MongoClient.Database(config.MongoDBName).Collection("ratings")

	filter := bson.M{"product_id": objID}
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		errs.Abort(c, errs.InternalError)
		return
	}

	var ratings []models.Rating
	if err = cursor.All(context.Background(), &ratings); err != nil {
		errs.Abort(c, errs.InternalError)
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
		errs.Abort(c, errs.ProductInvalidID)
		return
	}

	coll := config.MongoClient.Database(config.MongoDBName).Collection("ratings")
	var existing models.Rating
	err = coll.FindOne(c.Request.Context(), bson.M{"product_id": objID, "user_id": userID}).Decode(&existing)

	if err == mongo.ErrNoDocuments {
		c.JSON(http.StatusNotFound, gin.H{"code": "RATING_NOT_FOUND", "message": "You have not rated this product yet."})
		return
	}
	if err != nil {
		errs.Abort(c, errs.InternalError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"rating": existing.Rating, "id": existing.ID.Hex()})
}
