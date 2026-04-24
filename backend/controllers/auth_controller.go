package controllers

import (
	"log"
	"medieval-store/config"
	"medieval-store/errs"
	"medieval-store/models"
	"medieval-store/security"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SignupInput struct {
	Name        string  `json:"name" binding:"required"`
	Email       string  `json:"email" binding:"required,email"`
	TaxID       string  `json:"tax_id"`
	HomeAddress string  `json:"home_address"`
	Password    string  `json:"password" binding:"required,min=6"`
	Role        *string `json:"role,omitempty"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
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
		errs.Abort(c, errs.InternalError)
		log.Println("Failed to encrypt password")
		return
	}

	//Create user
	user := models.User{
		Name:        input.Name,
		Email:       input.Email,
		Password:    hashedPassword,
		TaxID:       input.TaxID,
		HomeAddress: input.HomeAddress,
		Role:        "customer",
	}

	if input.Role != nil {
		user.Role = *input.Role
	}

	var count int64
	if err := config.DB.Model(models.User{}).Where("email = ?", user.Email).Count(&count).Error; err != nil {
		errs.Abort(c, errs.InternalError)
		log.Printf("Failed to query email with error %s\n", err.Error())
		return
	}

	if count > 0 {
		errs.Abort(c, errs.AuthUserExists)
		return
	}

	//Save user to PostgreSQL database
	if err := config.DB.Create(&user).Error; err != nil {
		errs.Abort(c, errs.InternalError)
		log.Printf("Failed to create user with error %s\n", err.Error())
		return
	}

	user.Password = ""
	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully", "user": user})

}

func Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		errs.Abort(c, errs.AuthInvalidEmail)
		log.Printf("Failed login with error %s\n", err.Error())
		return
	}

	var user models.User

	//Find user
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		errs.Abort(c, errs.AuthInvalidCreds)
		return
	}

	//Verify password by checking its hash
	if match := security.CheckPasswordHash(input.Password, user.Password); !match {
		errs.Abort(c, errs.AuthInvalidCreds)
		return
	}

	//Generate JWT
	token, err := security.GenerateToken(user.ID, user.Role)
	if err != nil {
		errs.Abort(c, errs.InternalError)
		log.Println("Failed to generate token")
		return
	}

	//Send token back to frontend
	c.JSON(http.StatusOK, gin.H{"message": "Login successful", "token": token, "user_id": user.ID, "role": user.Role})
}
