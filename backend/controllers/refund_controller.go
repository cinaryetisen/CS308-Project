package controllers

import (
	"context"
	"net/http"
	"time"

	"medieval-store/config"
	"medieval-store/errs"
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
		errs.AbortWithDetail(c, errs.InvalidJSON, err.Error())
		return
	}

	// Fetch the order and verify ownership.
	var order models.Order
	if err := config.DB.First(&order, orderID).Error; err != nil {
		errs.Abort(c, errs.OrderNotFound)
		return
	}
	if order.CustomerID != userID.(uint) {
		errs.Abort(c, errs.OrderForbidden)
		return
	}

	// Order must be delivered before a refund can be requested.
	if order.Status != "delivered" {
		errs.Abort(c, errs.RefundIneligibleOrder)
		return
	}

	// Enforce the 30-day refund window.
	if time.Since(order.CreatedAt) > 30*24*time.Hour {
		errs.Abort(c, errs.RefundWindowExpired)
		return
	}

	// Fetch the order item and confirm it belongs to this order.
	var orderItem models.OrderItem
	if err := config.DB.First(&orderItem, input.OrderItemID).Error; err != nil {
		errs.AbortWithDetail(c, errs.OrderNotFound, "order item not found")
		return
	}
	if orderItem.OrderID != order.ID {
		errs.Abort(c, errs.RefundItemMismatch)
		return
	}

	// Prevent duplicate refund requests for the same item.
	var existing models.Refund
	err := config.DB.Where("order_item_id = ? AND status IN ?", input.OrderItemID, []string{"pending", "approved"}).
		First(&existing).Error
	if err == nil {
		errs.Abort(c, errs.RefundAlreadyExists)
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
		errs.Abort(c, errs.InternalError)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Refund request submitted successfully", "refund": refund})
}

// Returns all refund requests submitted by the logged-in customer.
func GetMyRefunds(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var refunds []models.Refund
	if err := config.DB.Where("customer_id = ?", userID).Find(&refunds).Error; err != nil {
		errs.Abort(c, errs.InternalError)
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
		errs.Abort(c, errs.InternalError)
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
		errs.AbortWithDetail(c, errs.InvalidJSON, err.Error())
		return
	}
	if input.Action != "approved" && input.Action != "rejected" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Action must be 'approved' or 'rejected'"})
		return
	}

	var refund models.Refund
	if err := config.DB.First(&refund, refundID).Error; err != nil {
		errs.Abort(c, errs.RefundNotFound)
		return
	}
	if refund.Status != "pending" {
		errs.Abort(c, errs.RefundAlreadyResolved)
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
			errs.Abort(c, errs.InternalError)
			return
		}

		objID, err := primitive.ObjectIDFromHex(orderItem.ProductID)
		if err != nil {
			tx.Rollback()
			errs.Abort(c, errs.ProductInvalidID)
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
			errs.Abort(c, errs.InternalError)
			return
		}

		// Mark the parent order as returned.
		if err := tx.Model(&models.Order{}).Where("id = ?", refund.OrderID).
			Update("status", "returned").Error; err != nil {
			tx.Rollback()
			errs.Abort(c, errs.InternalError)
			return
		}

		if err := tx.Save(&refund).Error; err != nil {
			tx.Rollback()
			errs.Abort(c, errs.InternalError)
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
			errs.Abort(c, errs.InternalError)
			return
		}

		var customer models.User
		if config.DB.First(&customer, refund.CustomerID).Error == nil {
			go services.SendRefundDecisionEmail(customer.Email, customer.Name, false, 0, refund.Reason)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Refund " + input.Action, "refund": refund})
}
