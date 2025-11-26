package controllers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sanjayk-eng/UserMenagmentSystem_Backend/models"
	"github.com/sanjayk-eng/UserMenagmentSystem_Backend/utils"
)

type EmployeeAuthData struct {
	ID       string `db:"id"`
	Email    string `db:"email"`
	Password string `db:"password"`
	Role     string `db:"role"`
	Status   string `db:"status"`
}

func (s *HandlerFunc) Login(c *gin.Context) {
	// 1. Parse request body
	var input models.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// 2. Fetch employee + role name
	emp, err := s.Query.GetEmployeeByEmail(input.Email)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, fmt.Sprintf("Login failed — email not found: %v", err.Error()))
		return
	}

	// 3. Validate password
	if !utils.CheckPassword(input.Password, emp.Password) {
		log.Printf("Login failed — wrong password for email: %s", input.Email)
		utils.RespondWithError(c, http.StatusUnauthorized, "Login failed — wrong password for email: "+input.Email)
		return
	}

	// 3.5 Check employee status
	if emp.Status == "deactive" {
		utils.RespondWithError(c, http.StatusForbidden, "Your account is deactivated. You cannot login")
		return
	}

	// 4. Generate JWT with role name
	token, err := utils.GenerateToken(emp.ID, emp.Role, s.Env.SERACT_KEY)
	if err != nil {
		log.Printf("JWT generation error: %v", err)
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to generate authentication token")
		return
	}

	// 5. Success Response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Login successful",
		"token":   token,
		"user": gin.H{
			"id":    emp.ID,
			"email": emp.Email,
			"role":  emp.Role,
		},
	})
}
