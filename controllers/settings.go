package controllers

import "github.com/gin-gonic/gin"

// GetCompanySettings - GET /api/settings/company
func GetCompanySettings(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Get company settings"})
}

// UpdateCompanySettings - POST /api/settings/company
func UpdateCompanySettings(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Update company settings"})
}

// GetPermissions - GET /api/settings/permissions
func GetPermissions(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Get permissions"})
}

// UpdatePermissions - POST /api/settings/permissions
func UpdatePermissions(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Update permissions"})
}
