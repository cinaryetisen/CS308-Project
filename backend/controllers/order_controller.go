package controllers

import (
	"context"
	"medieval-store/config"
	"medieval-store/models"
	"medieval-store/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetDeliveryList(c *gin.Context) {
	var orders []models.Order

	if err := config.DB.Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch deliveries"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validStatuses := map[string]bool{
		"processing": true,
		"in-transit": true,
		"delivered":  true,
	}
	if !validStatuses[input.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status. Must be processing, in-transit, or delivered"})
		return
	}

	var order models.Order
	if err := config.DB.First(&order, orderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	order.Status = input.Status

	if input.Status == "delivered" {
		order.Completed = true
	} else {
		order.Completed = false
	}

	if err := config.DB.Save(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order status updated successfully", "order": order})
}

func GetMyOrders(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var orders []models.Order

	if err := config.DB.Where("customer_id = ?", userID).Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch your orders"})
		return
	}

	c.JSON(http.StatusOK, orders)
}

func DownloadInvoice(c *gin.Context) {
	orderID := c.Param("id")
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	// 1. Fetch Order & preloaded Items
	var order models.Order
	if err := config.DB.Preload("Items").First(&order, orderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	// 2. SECURITY GUARD: Only the owner or a PM can download this invoice!
	if role != "product_manager" && order.CustomerID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to view this invoice"})
		return
	}

	// 3. Fetch User and Product names
	var user models.User
	config.DB.First(&user, order.CustomerID)

	var objIDs []primitive.ObjectID
	for _, item := range order.Items {
		id, _ := primitive.ObjectIDFromHex(item.ProductID)
		objIDs = append(objIDs, id)
	}

	productMap := make(map[string]models.Product)
	if len(objIDs) > 0 {
		collection := config.MongoClient.Database("medieval_store").Collection("products")
		cursor, _ := collection.Find(context.Background(), bson.M{"_id": bson.M{"$in": objIDs}})
		var products []models.Product
		cursor.All(context.Background(), &products)
		for _, p := range products {
			productMap[p.ID.Hex()] = p
		}
	}

	// 4. Generate PDF bytes on the fly
	pdfBytes, err := services.GenerateInvoicePDF(user, order, order.Items, productMap)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF"})
		return
	}

	// 5. Stream the raw bytes to the browser as a PDF!
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}
