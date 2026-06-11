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
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func AddToCart(c *gin.Context) {
	userID, _ := c.Get("user_id")

	//Bind the JSON from the React frontend
	var input struct {
		ProductID string `json:"product_id" binding:"required"`
		Quantity  int    `json:"quantity" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		errs.AbortWithDetail(c, errs.InvalidJSON, err.Error())
		return
	}

	item := models.CartItem{
		UserID:    userID.(uint),
		ProductID: input.ProductID,
		Quantity:  input.Quantity,
	}

	err := config.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "product_id"}}, // The composite primary keys
		DoUpdates: clause.Assignments(map[string]interface{}{
			"quantity": gorm.Expr("cart_items.quantity + ?", item.Quantity), // Increment the quantity safely
		}),
	}).Create(&item).Error

	if err != nil {
		errs.Abort(c, errs.InternalError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cart updated successfully"})
}

func GetCart(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var items []models.CartItem
	result := config.DB.Where("user_id = ?", userID).Find(&items)
	if result.Error != nil {
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
		c.JSON(200, []models.Product{})
		return
	}

	// Soft-deleted products silently drop out of the cart view.
	filter := bson.M{"_id": bson.M{"$in": objectIDs}, "deleted_at": nil}

	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	cursor, err := collection.Find(ctx, filter)
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

	var response []models.CartItemResponse
	for _, item := range items {
		if p, found := productMap[item.ProductID]; found {
			effective := p.EffectivePrice()
			response = append(response, models.CartItemResponse{
				ProductID:     p.ID.Hex(),
				Name:          p.Name,
				Price:         effective,
				OriginalPrice: p.Price,
				Discount:      p.Discount,
				Quantity:      item.Quantity,
				Subtotal:      float64(item.Quantity) * effective,
				ImageURL:      p.ImageURL,
				Stock:         p.Quantity,
				Category:      p.Category,
			})
		}
	}

	c.JSON(200, response)
}

func ClearCart(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	if err := config.DB.Where("user_id = ?", userID).Delete(&models.CartItem{}).Error; err != nil {
		errs.Abort(c, errs.InternalError)
		return
	}

	c.Status(204)
}

func RemoveFromCart(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	productID := c.Param("id")

	result := config.DB.Where("user_id = ? AND product_id = ?", userID, productID).Delete(&models.CartItem{})

	if result.Error != nil {
		errs.Abort(c, errs.InternalError)
		return
	}

	if result.RowsAffected == 0 {
		errs.Abort(c, errs.CartItemNotFound)
		return
	}

	c.Status(204)
}

func MergeCarts(c *gin.Context) {
	userId := c.MustGet("user_id").(uint)

	var guestItems []struct {
		ProductID string `json:"product_id" binding:"required"`
		Quantity  int    `json:"quantity" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&guestItems); err != nil {
		errs.Abort(c, errs.InvalidJSON)
		return
	}

	tx := config.DB.Begin()

	for _, item := range guestItems {
		item := models.CartItem{
			UserID:    userId,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}

		err := config.DB.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "user_id"}, {Name: "product_id"}}, // The composite primary keys
			DoUpdates: clause.Assignments(map[string]interface{}{
				"quantity": gorm.Expr("cart_items.quantity + ?", item.Quantity), // Increment the quantity safely
			}),
		}).Create(&item).Error

		if err != nil {
			tx.Rollback()
			errs.Abort(c, errs.InternalError)
			return
		}
	}

	tx.Commit()
	c.JSON(200, gin.H{"message": "Carts merged successfully"})
}
