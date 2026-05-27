package controllers

import (
	"net/http"

	"medieval-store/config"
	"medieval-store/errs"
	"medieval-store/models"

	"github.com/gin-gonic/gin"
)

// Function to get user information from database
func GetProfile(c *gin.Context) {
	//Extract the user_id from AuthMiddleware
	userID, exists := c.Get("user_id")
	if !exists {
		errs.Abort(c, errs.UserUnauthorized)
		return
	}

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		errs.Abort(c, errs.UserNotFound)
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
		errs.AbortWithDetail(c, errs.InvalidJSON, err.Error())
		return
	}

	//Fetch user from database
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		errs.Abort(c, errs.UserNotFound)
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
		errs.Abort(c, errs.InternalError)
		return
	}

	//Clear password hash before sending information to frontend
	user.Password = ""

	c.JSON(http.StatusOK, user)
}
