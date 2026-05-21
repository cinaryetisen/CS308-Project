package controllers

import (
	"context"
	"net/http"
	"time"

	"medieval-store/config"
	"medieval-store/models"
	"medieval-store/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

// Returns all refund requests submitted by the logged-in customer.
func GetMyRefunds(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var refunds []models.Refund
	if err := config.DB.Where("customer_id = ?", userID).Find(&refunds).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch your refund requests"})
		return
	}

	c.JSON(http.StatusOK, refunds)
}

// Returns refund requests filtered by status (default: pending).
// GET /api/admin/refunds?status=pending
func GetRefundRequests(c *gin.Context) {
	status := c.DefaultQuery("status", "pending")

	var refunds []models.Refund
	if err := config.DB.Where("status = ?", status).Find(&refunds).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch refund requests"})
		return
	}

	c.JSON(http.StatusOK, refunds)
}

// ResolveRefund approves or rejects a refund request.
// PATCH /api/admin/refunds/:id
// Approve path: restores stock in MongoDB and marks the order as "returned" inside a single transaction.
func ResolveRefund(c *gin.Context) {
	refundID := c.Param("id")
	managerID, _ := c.Get("user_id")

	var input struct {
		Action string `json:"action" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if input.Action != "approved" && input.Action != "rejected" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Action must be 'approved' or 'rejected'"})
		return
	}

	var refund models.Refund
	if err := config.DB.First(&refund, refundID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Refund request not found"})
		return
	}
	if refund.Status != "pending" {
		c.JSON(http.StatusConflict, gin.H{"error": "Refund has already been resolved"})
		return
	}

	now := time.Now()
	mid := managerID.(uint)
	refund.Status = input.Action
	refund.ResolvedAt = &now
	refund.ResolverID = &mid

	if input.Action == "approved" {
		// Wrap PostgreSQL updates + MongoDB stock restore in a single transaction so
		// a failed MongoDB call rolls back the status change too.
		tx := config.DB.Begin()

		var orderItem models.OrderItem
		if err := tx.First(&orderItem, refund.OrderItemID).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Order item not found"})
			return
		}

		objID, err := primitive.ObjectIDFromHex(orderItem.ProductID)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid product ID on order item"})
			return
		}

		// Restore stock atomically (mirrors checkout_controller.go:122).
		collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
		_, err = collection.UpdateOne(
			context.Background(),
			bson.M{"_id": objID},
			bson.M{"$inc": bson.M{"quantity": orderItem.Quantity}},
		)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to restore stock"})
			return
		}

		// Mark the parent order as returned.
		if err := tx.Model(&models.Order{}).Where("id = ?", refund.OrderID).
			Update("status", "returned").Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
			return
		}

		if err := tx.Save(&refund).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save refund"})
			return
		}

		tx.Commit()

		// Send approval email to the customer asynchronously.
		var customer models.User
		if config.DB.First(&customer, refund.CustomerID).Error == nil {
			go services.SendRefundDecisionEmail(customer.Email, customer.Name, true, refund.RefundAmount, "")
		}

	} else {
		// Rejection: no stock changes, no transaction needed.
		if err := config.DB.Save(&refund).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save refund"})
			return
		}

		var customer models.User
		if config.DB.First(&customer, refund.CustomerID).Error == nil {
			go services.SendRefundDecisionEmail(customer.Email, customer.Name, false, 0, refund.Reason)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Refund " + input.Action, "refund": refund})
}
