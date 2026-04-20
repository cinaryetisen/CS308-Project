package controllers

import (
	"net/http"

	"medieval-store/config"
	"medieval-store/models"

	"github.com/gin-gonic/gin"
)

// Function to get user information from database
func GetProfile(c *gin.Context) {
	//Extract the user_id from AuthMiddleware
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	//Clear password hash before sending information to frontend
	user.Password = ""

	c.JSON(http.StatusOK, user)
}

// Function to update user information (apart from email and password)
func UpdateProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")

	//Temporary struct to catch only the allowed fields
	var input struct {
		Name        string `json:"name"`
		TaxID       string `json:"tax_id"`
		HomeAddress string `json:"home_address"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//Fetch user from database
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	//Update the fields provided by the frontend
	if input.Name != "" {
		user.Name = input.Name
	}
	if input.TaxID != "" {
		user.TaxID = input.TaxID
	}
	if input.HomeAddress != "" {
		user.HomeAddress = input.HomeAddress
	}

	//Save changes to database
	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	//Clear password hash before sending information to frontend
	user.Password = ""

	c.JSON(http.StatusOK, user)
}
