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
	"gorm.io/gorm/clause"
)

// AddToWishlist is an idempotent insert operation into wishlist_items
func AddToWishlist(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var input struct {
		ProductID string `json:"product_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		errs.AbortWithDetail(c, errs.InvalidJSON, err.Error())
		return
	}

	item := models.WishlistItem{
		UserID:    userID,
		ProductID: input.ProductID,
	}

	if err := config.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&item).Error; err != nil {
		errs.Abort(c, errs.InternalError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Added to wishlist"})
}

// GetWishlist returns caller's wishlist with product details
func GetWishlist(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var items []models.WishlistItem
	if err := config.DB.Where("user_id = ?", userID).Find(&items).Error; err != nil {
		errs.Abort(c, errs.InternalError)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var objectIDs []primitive.ObjectID
	for _, i := range items {
		objID, err := primitive.ObjectIDFromHex(i.ProductID)
		if err == nil {
			objectIDs = append(objectIDs, objID)
		}
	}

	if len(objectIDs) == 0 {
		c.JSON(http.StatusOK, []models.WishlistItemResponse{})
		return
	}

	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	// Soft-deleted products silently drop out of the wishlist view.
	cursor, err := collection.Find(ctx, bson.M{"_id": bson.M{"$in": objectIDs}, "deleted_at": nil})
	if err != nil {
		errs.Abort(c, errs.InternalError)
		return
	}
	defer cursor.Close(ctx)

	var products []models.Product
	if err = cursor.All(ctx, &products); err != nil {
		errs.Abort(c, errs.InternalError)
		return
	}

	productMap := make(map[string]models.Product)
	for _, p := range products {
		productMap[p.ID.Hex()] = p
	}

	response := []models.WishlistItemResponse{}
	for _, item := range items {
		if p, found := productMap[item.ProductID]; found {
			response = append(response, models.WishlistItemResponse{
				ProductID: p.ID.Hex(),
				Name:      p.Name,
				Price:     p.Price,
				Discount:  p.Discount,
				ImageURL:  p.ImageURL,
				Stock:     p.Quantity,
				Category:  p.Category,
				AddedAt:   item.CreatedAt,
			})
		}
	}

	c.JSON(http.StatusOK, response)
}

func RemoveFromWishlist(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	productID := c.Param("productId")

	result := config.DB.Where("user_id = ? AND product_id = ?", userID, productID).Delete(&models.WishlistItem{})

	if result.Error != nil {
		errs.Abort(c, errs.InternalError)
		return
	}

	if result.RowsAffected == 0 {
		errs.Abort(c, errs.WishlistItemNotFound)
		return
	}

	c.Status(http.StatusNoContent)
}
