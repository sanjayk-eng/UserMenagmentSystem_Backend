package controllers

import "github.com/gin-gonic/gin"

// ApplyLeave - POST /api/leaves/apply
func ApplyLeave(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Apply leave"})
}

// AdminAddLeave - POST /api/leaves/admin-add
func AdminAddLeave(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Admin add leave"})
}

// ActionLeave - POST /api/leaves/:id/action
func ActionLeave(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Approve/Reject leave"})
}
