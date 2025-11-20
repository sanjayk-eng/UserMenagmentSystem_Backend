package controllers

import "github.com/gin-gonic/gin"

// Login - POST /api/auth/login
func Login(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Login endpoint"})
}
