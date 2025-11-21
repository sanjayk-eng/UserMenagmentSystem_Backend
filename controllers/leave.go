package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sanjayk-eng/UserMenagmentSystem_Backend/models"
	"github.com/sanjayk-eng/UserMenagmentSystem_Backend/utils"
)

// ApplyLeave - POST /api/leaves/apply
func (s *HandlerFunc) ApplyLeave(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Apply leave"})
}

// AdminAddLeave - POST /api/leaves/admin-add
func (s *HandlerFunc) AdminAddLeave(c *gin.Context) {
	roleValue, exists := c.Get("role")
	if !exists {
		utils.RespondWithError(c, http.StatusInternalServerError, "failed to get role")
		return
	}
	userRole := roleValue.(string)
	if userRole != "SUPERADMIN" {
		utils.RespondWithError(c, http.StatusUnauthorized, "not permitted to assign manager")
		return
	}
	var input models.LeaveTypeInput

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	// Set defaults
	if input.IsPaid == nil {
		defaultPaid := false
		input.IsPaid = &defaultPaid
	}
	if input.DefaultEntitlement == nil {
		defaultEntitlement := 0
		input.DefaultEntitlement = &defaultEntitlement
	}
	if input.LeaveCount == nil {
		defaultCount := 2
		input.LeaveCount = &defaultCount
	}

	if *input.LeaveCount <= 0 {
		utils.RespondWithError(c, http.StatusBadRequest, "leave_count must be greater than 0")
		return
	}

	query := `
		INSERT INTO Tbl_Leave_type (name, is_paid, default_entitlement)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`

	var leave models.LeaveType
	err := s.DB.QueryRow(query, input.Name, *input.IsPaid, *input.DefaultEntitlement).
		Scan(&leave.ID, &leave.CreatedAt, &leave.UpdatedAt)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to insert leave type: "+err.Error())
		return
	}

	leave.Name = input.Name
	leave.IsPaid = *input.IsPaid
	leave.DefaultEntitlement = *input.DefaultEntitlement

	c.JSON(http.StatusOK, leave)
}

func (s *HandlerFunc) GetAllLeavePolicies(c *gin.Context) {
	var leaves []models.LeaveType

	query := `SELECT id, name, is_paid, default_entitlement,  created_at, updated_at FROM Tbl_Leave_type ORDER BY id`
	err := s.DB.Select(&leaves, query)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch leave types: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, leaves) // send models directly
}

// ActionLeave - POST /api/leaves/:id/action
func (s *HandlerFunc) ActionLeave(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Approve/Reject leave"})
}
