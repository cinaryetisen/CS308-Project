package controllers

import (
	"medieval-store/config"
	"medieval-store/models"
	"medieval-store/security"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SignupInput struct {
	Name        string `json:"name" binding:"required"`
	Email       string `json:"email" binding:"required, email"`
	TaxID       string `json:"tax_id"`
	HomeAddress string `json:"home_address"`
	Password    string `json:"password" binding:"required, min=6"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required, email"`
	Password string `json:"password" binding:"required, min=6"`
}

func Signup(c *gin.Context) {
	var input SignupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//Hash the password
	hashedPassword, err := security.HashPassword(input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt password"})
		return
	}

	//Create user
	user := models.User{
		Name:        input.Name,
		Email:       input.Email,
		Password:    hashedPassword,
		TaxID:       input.TaxID,
		HomeAddress: input.HomeAddress,
	}

	//Save user to PostgreSQL database
	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user. Email might already exist."})
	}

	user.Password = ""
	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully", "user": user})

}

func Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User

	//Find user
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	//Verify password by checking its hash
	if match := security.CheckPasswordHash(input.Password, user.Password); !match {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Login successful", "user_id": user.ID, "role": user.Role})
}
