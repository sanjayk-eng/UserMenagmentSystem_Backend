package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sanjayk-eng/UserMenagmentSystem_Backend/models"
	"github.com/sanjayk-eng/UserMenagmentSystem_Backend/utils"
	"github.com/sanjayk-eng/UserMenagmentSystem_Backend/utils/common"
	"github.com/sanjayk-eng/UserMenagmentSystem_Backend/utils/constant"
)

// GetCompanySettings - GET /api/settings/company
func (h *HandlerFunc) GetCompanySettings(c *gin.Context) {
	// Only SUPERADMIN and ADMIN allowed
	roleRaw, _ := c.Get("role")
	role := roleRaw.(string)
	if role != "SUPERADMIN" && role != "ADMIN" && role != "HR" {
		utils.RespondWithError(c, 403, "Not authorized to view settings")
		return
	}
	var settings models.CompanySettings
	err := h.Query.GetCompanySettings(settings)
	if err != nil {
		utils.RespondWithError(c, 500, "Failed to fetch settings: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"settings": settings,
	})
}

// UpdateCompanySettings - PUT /api/settings/company
func (h *HandlerFunc) UpdateCompanySettings(c *gin.Context) {
	// Only SUPERADMIN and ADMIN allowed
	roleRaw, _ := c.Get("role")
	role := roleRaw.(string)
	if role != "SUPERADMIN" && role != "ADMIN" && role != "HR" {
		utils.RespondWithError(c, 403, "Not authorized to update settings")
		return
	}
	var input models.CompanyField

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, 400, "Invalid input: "+err.Error())
		return
	}
	empIDRaw, ok := c.Get("user_id")
	if !ok {
		utils.RespondWithError(c, http.StatusUnauthorized, "Employee ID missing")
		return

	}

	empIDStr, ok := empIDRaw.(string)
	if !ok {
		utils.RespondWithError(c, http.StatusInternalServerError, "Invalid employee ID format")
		return
	}

	empID, err := uuid.Parse(empIDStr)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Invalid employee UUID")
		return
	}

	err = common.ExecuteTransaction(c, h.Query.DB, func(tx *sqlx.Tx) error {
		err := h.Query.UpdateCompanySettings(tx, input)
		if err != nil {
			return utils.CustomErr(c, 500, "Failed to fetch settings: "+err.Error())
		}
		//add log
		data := utils.NewCommon(constant.CompanySettings, constant.ActionCreate, empID)

		err = common.AddLog(data, tx)
		if err != nil {
			return utils.CustomErr(c, http.StatusInternalServerError, "Failed to log action: "+err.Error())
		}
		return err
	})

	if err != nil {
		utils.RespondWithError(c, 500, "Failed to update settings: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Company settings updated successfully",
	})
}
