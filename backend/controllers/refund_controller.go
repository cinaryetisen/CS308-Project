package controllers

import (
	"net/http"
	"time"

	"medieval-store/config"
	"medieval-store/models"

	"github.com/gin-gonic/gin"
)

func RequestRefund(c *gin.Context) {
	orderID := c.Param("id")
	userID, _ := c.Get("user_id")

	var input struct {
		OrderItemID uint   `json:"order_item_id" binding:"required"`
		Reason      string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Fetch the order and verify ownership.
	var order models.Order
	if err := config.DB.First(&order, orderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}
	if order.CustomerID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "This order does not belong to you"})
		return
	}

	// Order must be delivered before a refund can be requested.
	if order.Status != "delivered" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only delivered orders are eligible for a refund"})
		return
	}

	// Enforce the 30-day refund window.
	if time.Since(order.CreatedAt) > 30*24*time.Hour {
		c.JSON(http.StatusBadRequest, gin.H{"error": "The 30-day refund window for this order has passed"})
		return
	}

	// Fetch the order item and confirm it belongs to this order.
	var orderItem models.OrderItem
	if err := config.DB.First(&orderItem, input.OrderItemID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order item not found"})
		return
	}
	if orderItem.OrderID != order.ID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order item does not belong to this order"})
		return
	}

	// Prevent duplicate refund requests for the same item.
	var existing models.Refund
	err := config.DB.Where("order_item_id = ? AND status IN ?", input.OrderItemID, []string{"pending", "approved"}).
		First(&existing).Error
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "A refund request already exists for this item"})
		return
	}

	// RefundAmount uses the purchase-time price snapshot (OrderItem.Price), which already
	// reflects any discount that was active at checkout.
	refund := models.Refund{
		OrderID:      order.ID,
		OrderItemID:  input.OrderItemID,
		CustomerID:   userID.(uint),
		RefundAmount: orderItem.Price * float64(orderItem.Quantity),
		Reason:       input.Reason,
		Status:       "pending",
	}

	if err := config.DB.Create(&refund).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit refund request"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Refund request submitted successfully", "refund": refund})
}
