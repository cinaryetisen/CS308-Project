package controllers

import (
	"context"
	"log"
	"net/http"

	"medieval-store/config"
	"medieval-store/models"
	"medieval-store/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Checkout processes the order, securely deducts stock, and emails a PDF receipt
func Checkout(c *gin.Context) {
	// 1. Get the logged-in user's ID from the JWT middleware
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized. Please log in to checkout."})
		return
	}

	// 2. Bind the incoming checkout data from React
	var input struct {
		ShippingAddress string            `json:"shipping_address" binding:"required"`
		TotalPrice      float64           `json:"total_price" binding:"required"`
		CartItems       []models.CartItem `json:"cart_items" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid checkout data: " + err.Error()})
		return
	}

	// ==========================================
	// SECURE PRICING & NAMES: Fetch full product data from MongoDB
	// ==========================================
	var objectIDs []primitive.ObjectID
	for _, item := range input.CartItems {
		objID, err := primitive.ObjectIDFromHex(item.ProductID)
		if err == nil {
			objectIDs = append(objectIDs, objID)
		}
	}

	productMap := make(map[string]models.Product)

	if len(objectIDs) > 0 {
		collection := config.MongoClient.Database("medieval_store").Collection("products")
		cursor, err := collection.Find(context.Background(), bson.M{"_id": bson.M{"$in": objectIDs}})

		if err == nil {
			var products []models.Product
			if err = cursor.All(context.Background(), &products); err == nil {
				for _, p := range products {
					productMap[p.ID.Hex()] = p
				}
			}
		}
	}
	// ==========================================

	// 3. Create the new order in the database
	newOrder := models.Order{
		CustomerID:      userID.(uint),
		TotalPrice:      input.TotalPrice,
		DeliveryAddress: input.ShippingAddress,
		Status:          "processing",
		Completed:       false,
	}

	// Start a Database Transaction for safety
	tx := config.DB.Begin()

	if err := tx.Create(&newOrder).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	// 4. Process individual items & Deduct Stock
	for _, item := range input.CartItems {

		// Extract secure price
		verifiedPrice := 0.00
		if p, found := productMap[item.ProductID]; found {
			verifiedPrice = p.Price
		}

		orderItem := models.OrderItem{
			OrderID:   newOrder.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     verifiedPrice,
		}

		if err := tx.Create(&orderItem).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save order items"})
			return
		}

		// ==========================================
		// ATOMIC CONCURRENCY SHIELD: Deduct Stock
		// ==========================================
		objID, _ := primitive.ObjectIDFromHex(item.ProductID)
		productsCollection := config.MongoClient.Database("medieval_store").Collection("products")

		// Filter: Only match the product IF it has enough stock (>= item.Quantity)
		filter := bson.M{
			"_id":      objID,
			"quantity": bson.M{"$gte": item.Quantity},
		}

		// Update: Subtract the amount
		update := bson.M{"$inc": bson.M{"quantity": -item.Quantity}}

		result, err := productsCollection.UpdateOne(context.Background(), filter, update)

		// If 0 documents were modified, it means another user bought the last item right before this!
		if err != nil || result.ModifiedCount == 0 {
			tx.Rollback() // Cancel the PostgreSQL order completely
			c.JSON(http.StatusConflict, gin.H{
				"error": "Checkout failed. One or more items in your cart just went out of stock!",
			})
			return
		}
		// ==========================================
	}

	// 5. Clear the user's shopping cart
	if err := tx.Where("user_id = ?", userID).Delete(&models.CartItem{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear shopping cart"})
		return
	}

	// 6. Permanently commit the PostgreSQL changes!
	tx.Commit()

	// 7. Fetch the User's details and trigger the background PDF logic
	var user models.User
	if err := config.DB.First(&user, userID).Error; err == nil {

		go func(u models.User, order models.Order, items []models.CartItem, pMap map[string]models.Product) {
			log.Println("Generating PDF Invoice...")

			pdfPath, err := services.GenerateInvoicePDF(u, order, items, pMap)
			if err != nil {
				log.Printf("Error generating PDF: %v\n", err)
				return
			}

			log.Println("PDF Generated successfully. Sending email...")

			err = services.SendInvoiceEmail(u.Email, pdfPath)
			if err != nil {
				log.Printf("Error sending invoice email: %v\n", err)
			} else {
				log.Printf("Invoice successfully sent to %s\n", u.Email)
			}

		}(user, newOrder, input.CartItems, productMap)
	}

	// 8. Instantly send success response to frontend
	c.JSON(http.StatusCreated, gin.H{
		"message": "Order placed successfully! Your receipt is being dispatched.",
		"order":   newOrder,
	})
}
