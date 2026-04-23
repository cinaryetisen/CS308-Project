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

// Checkout processes the order, generates a PDF invoice, and emails it
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

	var objectIDs []primitive.ObjectID
	for _, item := range input.CartItems {
		objID, err := primitive.ObjectIDFromHex(item.ProductID)
		if err == nil {
			objectIDs = append(objectIDs, objID)
		}
	}

	// Create a map to hold the ENTIRE product struct (for prices AND names)
	productMap := make(map[string]models.Product)

	if len(objectIDs) > 0 {
		collection := config.MongoClient.Database("medieval_store").Collection("products")
		cursor, err := collection.Find(context.Background(), bson.M{"_id": bson.M{"$in": objectIDs}})

		if err == nil {
			var products []models.Product
			if err = cursor.All(context.Background(), &products); err == nil {
				// Populate the map: Key = ProductID string, Value = Full Product Struct
				for _, p := range products {
					productMap[p.ID.Hex()] = p
				}
			}
		}
	}

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

	// 3a. Save the main Order
	if err := tx.Create(&newOrder).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	// 3b. Loop through cart items and attach them to the Order
	for _, item := range input.CartItems {

		// Securely extract the price from our MongoDB map
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
	}

	// 3c. Clear the user's shopping cart
	if err := tx.Where("user_id = ?", userID).Delete(&models.CartItem{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear shopping cart"})
		return
	}

	// 3d. If everything succeeded, permanently commit the changes!
	tx.Commit()

	// 4. Fetch the User's details
	var user models.User
	if err := config.DB.First(&user, userID).Error; err == nil {

		// 5. Run the PDF & Email logic in a BACKGROUND goroutine
		go func(u models.User, order models.Order, items []models.CartItem, pMap map[string]models.Product) {
			log.Println("Generating PDF Invoice...")

			// Generate the PDF
			pdfPath, err := services.GenerateInvoicePDF(u, order, items, pMap)
			if err != nil {
				log.Printf("Error generating PDF: %v\n", err)
				return
			}

			log.Println("PDF Generated successfully. Sending email...")

			// Email the PDF
			err = services.SendInvoiceEmail(u.Email, pdfPath)
			if err != nil {
				log.Printf("Error sending invoice email: %v\n", err)
			} else {
				log.Printf("Invoice successfully sent to %s\n", u.Email)
			}

		}(user, newOrder, input.CartItems, productMap)
	}

	// 6. Instantly send success response to frontend
	c.JSON(http.StatusCreated, gin.H{
		"message": "Order placed successfully! Your receipt is being dispatched.",
		"order":   newOrder,
	})
}
