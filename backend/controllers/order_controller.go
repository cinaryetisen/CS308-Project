package controllers

import (
	"context"
	"medieval-store/config"
	"medieval-store/errs"
	"medieval-store/models"
	"medieval-store/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetDeliveryList(c *gin.Context) {
	var orders []models.Order

	if err := config.DB.Preload("Items").Find(&orders).Error; err != nil {
		errs.Abort(c, errs.InternalError)
		return
	}

	c.JSON(http.StatusOK, orders)
}

func UpdateOrderStatus(c *gin.Context) {
	orderID := c.Param("id")

	var input struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		errs.AbortWithDetail(c, errs.InvalidJSON, err.Error())
		return
	}

	validStatuses := map[string]bool{
		"processing": true,
		"in-transit": true,
		"delivered":  true,
	}
	if !validStatuses[input.Status] {
		errs.Abort(c, errs.OrderInvalidStatus)
		return
	}

	var order models.Order
	if err := config.DB.First(&order, orderID).Error; err != nil {
		errs.Abort(c, errs.OrderNotFound)
		return
	}

	order.Status = input.Status

	if input.Status == "delivered" {
		order.Completed = true
	} else {
		order.Completed = false
	}

	if err := config.DB.Save(&order).Error; err != nil {
		errs.Abort(c, errs.InternalError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order status updated successfully", "order": order})
}

func GetMyOrders(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		errs.Abort(c, errs.UserUnauthorized)
		return
	}

	var orders []models.Order

	if err := config.DB.Preload("Items").Where("customer_id = ?", userID).Find(&orders).Error; err != nil {
		errs.Abort(c, errs.InternalError)
		return
	}

	c.JSON(http.StatusOK, orders)
}

// CancelOrder lets a customer cancel their own order while it is still in "processing". Restocks every line item in MongoDB and flips the order status to "cancelled". In-transit and delivered orders must go through refund flow
func CancelOrder(c *gin.Context) {
	orderID := c.Param("id")
	userID, _ := c.Get("user_id")

	var order models.Order
	if err := config.DB.Preload("Items").First(&order, orderID).Error; err != nil {
		errs.Abort(c, errs.OrderNotFound)
		return
	}

	if order.CustomerID != userID.(uint) {
		errs.Abort(c, errs.OrderForbidden)
		return
	}

	if order.Status != "processing" {
		errs.Abort(c, errs.OrderNotCancellable)
		return
	}

	tx := config.DB.Begin()

	//Restock each item atomically
	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	for _, item := range order.Items {
		objId, err := primitive.ObjectIDFromHex(item.ProductID)
		if err != nil {
			tx.Rollback()
			errs.Abort(c, errs.ProductInvalidID)
			return
		}
		if _, err := collection.UpdateOne(
			context.Background(),
			bson.M{"_id": objId},
			bson.M{"$inc": bson.M{"quantity": item.Quantity}},
		); err != nil {
			tx.Rollback()
			errs.Abort(c, errs.InternalError)
			return
		}
	}

	if err := tx.Model(&models.Order{}).Where("id = ?", order.ID).Update("status", "cancelled").Error; err != nil {
		tx.Rollback()
		errs.Abort(c, errs.InternalError)
		return
	}

	tx.Commit()

	order.Status = "cancelled"
	c.JSON(http.StatusOK, gin.H{"message": "Order cancelled successfully", "order": order})
}

func DownloadInvoice(c *gin.Context) {
	orderID := c.Param("id")
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	// 1. Fetch Order & preloaded Items
	var order models.Order
	if err := config.DB.Preload("Items").First(&order, orderID).Error; err != nil {
		errs.Abort(c, errs.OrderNotFound)
		return
	}

	// 2. SECURITY GUARD: Only the owner or a manager can download this invoice.
	// Sales managers need access too — req. 11 has them viewing/printing all invoices.
	if role != "product_manager" && role != "sales_manager" && order.CustomerID != userID.(uint) {
		errs.Abort(c, errs.OrderForbidden)
		return
	}

	// 3. Fetch User and Product names
	var user models.User
	if err := config.DB.First(&user, order.CustomerID).Error; err != nil {
		errs.Abort(c, errs.InternalError)
		return
	}

	var objIDs []primitive.ObjectID
	for _, item := range order.Items {
		id, err := primitive.ObjectIDFromHex(item.ProductID)
		if err == nil {
			objIDs = append(objIDs, id)
		}
	}

	productMap := make(map[string]models.Product)
	if len(objIDs) > 0 {
		collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
		cursor, err := collection.Find(context.Background(), bson.M{"_id": bson.M{"$in": objIDs}})
		if err != nil {
			errs.Abort(c, errs.InternalError)
			return
		}
		var products []models.Product
		if err := cursor.All(context.Background(), &products); err != nil {
			errs.Abort(c, errs.InternalError)
			return
		}
		for _, p := range products {
			productMap[p.ID.Hex()] = p
		}
	}

	// 4. Generate PDF bytes on the fly
	pdfBytes, err := services.GenerateInvoicePDF(user, order, order.Items, productMap)
	if err != nil {
		errs.Abort(c, errs.InternalError)
		return
	}

	// 5. Stream the raw bytes to the browser as a PDF!
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}
