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
