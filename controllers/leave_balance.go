package controllers

import "github.com/gin-gonic/gin"

// GetLeaveBalances - GET /api/employees/:id/leave-balances
func GetLeaveBalances(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Get leave balances"})
}

// AdjustLeaveBalance - POST /api/leave-balances/:id/adjust
func AdjustLeaveBalance(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Adjust leave balance"})
}
