package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sanjayk-eng/UserMenagmentSystem_Backend/utils"
	"github.com/sanjayk-eng/UserMenagmentSystem_Backend/utils/common"
	"github.com/sanjayk-eng/UserMenagmentSystem_Backend/utils/constant"
)

// AddHoliday handles adding a new holiday
func (s *HandlerFunc) AddHoliday(c *gin.Context) {
	role, _ := c.Get("role")
	if role.(string) != "SUPERADMIN" {
		utils.RespondWithError(c, http.StatusUnauthorized, "not permitted")
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

	var input struct {
		Name string    `json:"name" binding:"required"`
		Date time.Time `json:"date" binding:"required"`
		Type string    `json:"type"` // optional, defaults to HOLIDAY
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	// Normalize date to UTC midnight to avoid timezone issues
	normalizedDate := time.Date(
		input.Date.Year(),
		input.Date.Month(),
		input.Date.Day(),
		0, 0, 0, 0,
		time.UTC,
	)
	var holidayId string

	err = common.ExecuteTransaction(c, s.Query.DB, func(tx *sqlx.Tx) error {
		id, err := s.Query.AddHoliday(tx, input.Name, normalizedDate, input.Type)
		if err != nil {
			// utils.RespondWithError(c, http.StatusInternalServerError, "Failed to add holiday: "+err.Error())
			return utils.CustomErr(c, http.StatusInternalServerError, "Failed to add holiday: "+err.Error())
		}
		holidayId = id
		data := utils.NewCommon(constant.ComponentHoliday, constant.ActionCreate, empID)

		err = common.AddLog(data, tx)
		if err != nil {
			// utils.RespondWithError(c, http.StatusInternalServerError, "Failed to log action: "+err.Error())
			return utils.CustomErr(c, http.StatusInternalServerError, "Failed to log action: "+err.Error())
		}

		return err
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Holiday added successfully",
		"id":      holidayId,
		"date":    normalizedDate.Format("2006-01-02"),
	})
}

// GetHolidays returns all holidays
func (s *HandlerFunc) GetHolidays(c *gin.Context) {
	// role, _ := c.Get("role")
	// if role.(string) != "SUPERADMIN" {
	// 	utils.RespondWithError(c, http.StatusUnauthorized, "not permitted")
	// 	return
	// }

	holidays, err := s.Query.GetAllHolidays()
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch holidays: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, holidays)
}

// DeleteHoliday removes a holiday
func (s *HandlerFunc) DeleteHoliday(c *gin.Context) {
	role, _ := c.Get("role")
	if role.(string) != "SUPERADMIN" {
		utils.RespondWithError(c, http.StatusUnauthorized, "not permitted")
		return
	}

	id := c.Param("id")
	if id == "" {
		utils.RespondWithError(c, http.StatusBadRequest, "Holiday ID is required")
		return
	}

	err := common.ExecuteTransaction(c, s.Query.DB, func(tx *sqlx.Tx) error {
		err := s.Query.DeleteHoliday(id, tx)
		if err != nil {
			//utils.RespondWithError(c, http.StatusInternalServerError, "Failed to delete holiday: "+err.Error())
			return utils.CustomErr(c, http.StatusInternalServerError, "Failed to delete holiday: "+err.Error())

		}

		//add log
		empIDRaw, ok := c.Get("user_id")
		if !ok {
			return utils.CustomErr(c, http.StatusUnauthorized, "Employee ID missing")

		}

		empIDStr, ok := empIDRaw.(string)
		if !ok {
			return utils.CustomErr(c, http.StatusInternalServerError, "Invalid employee ID format")
		}

		empID, err := uuid.Parse(empIDStr)
		if err != nil {
			return utils.CustomErr(c, http.StatusInternalServerError, "Invalid employee UUID")
		}
		data := utils.NewCommon(constant.ComponentHoliday, constant.ActionDelete, empID)

		err = common.AddLog(data, tx)
		if err != nil {
			return utils.CustomErr(c, http.StatusInternalServerError, "Failed to log action: "+err.Error())
		}
		return err
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Holiday deleted successfully",
	})
}
