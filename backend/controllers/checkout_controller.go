package controllers

import (
	"log"
	"net/http"

	"medieval-store/config"
	"medieval-store/models"
	"medieval-store/services"

	"github.com/gin-gonic/gin"
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

	// 3. Create the new order in the database
	newOrder := models.Order{
		CustomerID: userID.(uint),
		Status:     "processing",
		Completed:  false,
	}

	if err := config.DB.Create(&newOrder).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	// 4. Fetch the User's details
	var user models.User
	if err := config.DB.First(&user, userID).Error; err == nil {

		// 5. Run the PDF & Email logic in a BACKGROUND goroutine
		go func(u models.User, order models.Order, items []models.CartItem) {
			log.Println("Generating PDF Invoice...")

			// 1. Generate the PDF
			pdfPath, err := services.GenerateInvoicePDF(u, order, items)
			if err != nil {
				log.Printf("Error generating PDF: %v\n", err)
				return // Stop if PDF fails
			}

			log.Println("PDF Generated successfully. Sending email...")

			// 2. Email the PDF
			err = services.SendInvoiceEmail(u.Email, pdfPath)
			if err != nil {
				log.Printf("Error sending invoice email: %v\n", err)
			} else {
				log.Printf("Invoice successfully sent to %s\n", u.Email)
			}

		}(user, newOrder, input.CartItems)
	}

	// 6. Instantly send success response to frontend
	c.JSON(http.StatusCreated, gin.H{
		"message": "Order placed successfully! Your receipt is being dispatched via raven.",
		"order":   newOrder,
	})
}
