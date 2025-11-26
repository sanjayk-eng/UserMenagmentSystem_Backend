package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sanjayk-eng/UserMenagmentSystem_Backend/utils"
)

// AddHoliday handles adding a new holiday
func (s *HandlerFunc) AddHoliday(c *gin.Context) {
	role, _ := c.Get("role")
	if role.(string) != "SUPERADMIN" {
		utils.RespondWithError(c, http.StatusUnauthorized, "not permitted")
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

	id, err := s.Query.AddHoliday(input.Name, input.Date, input.Type)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to add holiday: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Holiday added successfully",
		"id":      id,
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

	err := s.Query.DeleteHoliday(id)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to delete holiday: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Holiday deleted successfully",
	})
}
