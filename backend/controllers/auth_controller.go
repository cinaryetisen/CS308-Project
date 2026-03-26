package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type SignupInput struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required, email"`
	TaxID    int32  `json:"tax_id"`
	Address  string `json:"home_address"`
	Password string `json:"password" binding:"required, min=6"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required, email"`
	Password string `json:"password" binding:"required, min=6"`
}

func Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
}

func Signup(c *gin.Context) {
	var input SignupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
}
