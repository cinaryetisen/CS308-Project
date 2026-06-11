package controllers

import (
	"errors"
	"log"
	"medieval-store/config"
	"medieval-store/errs"
	"medieval-store/models"
	"medieval-store/security"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type SignupInput struct {
	Name        string `json:"name" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	TaxID       string `json:"tax_id"`
	HomeAddress string `json:"home_address"`
	Password    string `json:"password" binding:"required,min=6"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

func Signup(c *gin.Context) {
	var input SignupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		log.Printf("Signup Input JSON parsing failed with error: %s\n", err.Error())
		var verrs validator.ValidationErrors
		if errors.As(err, &verrs) {
			for _, fe := range verrs {
				switch fe.Field() {
				case "Password":
					// min=6 (too short) or required (missing)
					errs.Abort(c, errs.AuthWeakPassword)
					return
				case "Email":
					errs.Abort(c, errs.AuthInvalidEmail)
					return
				case "Name":
					errs.AbortWithDetail(c, errs.InvalidJSON, "name is required")
					return
				}
			}
		}
		// Malformed JSON or any unrecognised validation failure.
		errs.Abort(c, errs.InvalidJSON)
		return
	}

	//Hash the password
	hashedPassword, err := security.HashPassword(input.Password)
	if err != nil {
		errs.Abort(c, errs.InternalError)
		log.Println("Failed to encrypt password")
		return
	}

	//Create user (public signup is always a customer)
	user := models.User{
		Name:        input.Name,
		Email:       input.Email,
		Password:    hashedPassword,
		TaxID:       input.TaxID,
		HomeAddress: input.HomeAddress,
		Role:        "customer",
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
