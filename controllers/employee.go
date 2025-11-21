package controllers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sanjayk-eng/UserMenagmentSystem_Backend/models"
	"github.com/sanjayk-eng/UserMenagmentSystem_Backend/utils"
)

type UpdateRoleInput struct {
	Role string `json:"role" validate:"required"` // Only valid roles
}
type UpdateManagerInput struct {
	ManagerID string `json:"manager_id" validate:"required"` // UUID of new manager
}

// GetEmployees - GET /api/employees
func (h *HandlerFunc) GetEmployee(c *gin.Context) {
	roleValue, exist := c.Get("role")
	if !exist {
		utils.RespondWithError(c, http.StatusInternalServerError, "failed to get role")
		return
	}

	role := roleValue.(string)
	if role != "SUPERADMIN" && role != "Admin" {
		utils.RespondWithError(c, http.StatusUnauthorized, "not permit")
		return
	}
	rows, err := h.DB.Query(`
        SELECT 
            e.id, e.full_name, e.email,e.status, r.type as role,
            e.password, e.manager_id, e.salary, e.joining_date,
            e.created_at, e.updated_at, e.deleted_at
        FROM Tbl_Employee e
        JOIN Tbl_Role r ON e.role_id = r.id
        ORDER BY e.full_name
    `)
	if err != nil {
		utils.RespondWithError(c, 500, fmt.Sprintf("failed to fetch employees: %v", err.Error()))
		return
	}
	defer rows.Close()

	var employees []models.EmployeeInput

	for rows.Next() {
		var emp models.EmployeeInput
		err := rows.Scan(
			&emp.ID,
			&emp.FullName,
			&emp.Email,
			&emp.Status,
			&emp.Role,
			&emp.Password,
			&emp.ManagerID,
			&emp.Salary,
			&emp.JoiningDate,
			&emp.CreatedAt,
			&emp.UpdatedAt,
			&emp.DeletedAt,
		)
		if err != nil {
			utils.RespondWithError(c, 500, fmt.Sprintf("failed to read employee data: %v", err.Error()))
			return
		}

		employees = append(employees, emp)
	}

	c.JSON(200, gin.H{
		"message":   "All employees fetched successfully",
		"employees": employees,
	})
}

func (h *HandlerFunc) GetEmployeeById(c *gin.Context) {

}

func (h *HandlerFunc) CreateEmployee(c *gin.Context) {
	roleValue, exist := c.Get("role")
	if !exist {
		utils.RespondWithError(c, http.StatusInternalServerError, "failed to get role")
		return
	}

	role := roleValue.(string)
	if role != "SUPERADMIN" && role != "Admin" {
		utils.RespondWithError(c, http.StatusUnauthorized, "not permit")
		return
	}
	var input models.EmployeeInput

	// Bind JSON input
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, 400, err.Error())
		return
	}

	// Standard validator check
	if err := models.Validate.Struct(input); err != nil {
		utils.RespondWithError(c, 400, err.Error())
		return
	}

	// Enforce @zenithive.com email suffix
	if !strings.HasSuffix(input.Email, "@zenithive.com") {
		utils.RespondWithError(c, 400, "email must end with @zenithive.com")
		return
	}

	// Check if email already exists
	var existingEmail string
	err := h.DB.QueryRow(`SELECT email FROM Tbl_Employee WHERE email=$1`, input.Email).Scan(&existingEmail)
	if err == nil {
		utils.RespondWithError(c, 400, "email already exists")
		return
	} else if err != sql.ErrNoRows {
		utils.RespondWithError(c, 500, "database error")
		return
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		utils.RespondWithError(c, 500, "failed to hash password")
		return
	}

	// Get role_id from role type
	var id string
	err = h.DB.QueryRow(`SELECT id FROM Tbl_Role WHERE type=$1`, input.Role).Scan(&id)
	if err != nil {
		utils.RespondWithError(c, 400, fmt.Sprintf("role not found %s", err.Error()))
		return
	}

	// Insert into DB
	_, err = h.DB.Exec(`
        INSERT INTO Tbl_Employee (full_name, email, role_id, password, salary, joining_date)
        VALUES ($1, $2, $3, $4, $5, $6)`,
		input.FullName, input.Email, id, hashedPassword, input.Salary, input.JoiningDate,
	)
	if err != nil {
		utils.RespondWithError(c, 500, "failed to create employee")
		return
	}

	// Success response
	c.JSON(201, gin.H{"message": "employee created successfully"})
}

func (s *HandlerFunc) UpdateEmployeeRole(c *gin.Context) {
	// 1. Get current user's role
	roleValue, exists := c.Get("role")
	if !exists {
		utils.RespondWithError(c, http.StatusInternalServerError, "failed to get role")
		return
	}
	userRole := roleValue.(string)

	// 2. Only SUPERADMIN / ADMIN can update roles
	if userRole != "SUPERADMIN" && userRole != "ADMIN" {
		utils.RespondWithError(c, http.StatusUnauthorized, "not permitted to update role")
		return
	}

	// 3. Get target employee ID
	empID := c.Param("id")
	if empID == "" {
		utils.RespondWithError(c, http.StatusBadRequest, "employee_id is required")
		return
	}

	// 4. Bind input
	var input UpdateRoleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "invalid JSON")
		return
	}

	// 5. Validate role
	validRoles := map[string]bool{
		"SUPERADMIN": true,
		"ADMIN":      true,
		"HR":         true,
		"EMPLOYEE":   true,
		"MANAGAR":    true,
	}
	if !validRoles[input.Role] {
		utils.RespondWithError(c, http.StatusBadRequest, "invalid role")
		return
	}

	// 6. Check if the employee already has this role
	var currentRole string
	err := s.DB.QueryRow(`
        SELECT R.TYPE
        FROM TBL_EMPLOYEE E
        JOIN TBL_ROLE R ON E.ROLE_ID = R.ID
        WHERE E.ID = $1
    `, empID).Scan(&currentRole)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "failed to fetch current role")
		return
	}

	if currentRole == input.Role {
		// Role is the same, no update needed
		c.JSON(http.StatusOK, gin.H{
			"message":     "employee already has this role",
			"employee_id": empID,
			"role":        currentRole,
		})
		return
	}

	// 7. Update role
	query := `
        UPDATE TBL_EMPLOYEE
        SET ROLE_ID = (SELECT ID FROM TBL_ROLE WHERE TYPE=$1),
            UPDATED_AT = NOW()
        WHERE ID = $2
        RETURNING ID;
    `
	var updatedID string
	err = s.DB.QueryRow(query, input.Role, empID).Scan(&updatedID)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	// 8. Success response
	c.JSON(http.StatusOK, gin.H{
		"message":     "employee role updated successfully",
		"employee_id": updatedID,
		"new_role":    input.Role,
	})
}
func (s *HandlerFunc) UpdateEmployeeManager(c *gin.Context) {
	// 1. Role check (only HR, SUPERADMIN, ADMIN can assign)
	roleValue, exists := c.Get("role")
	if !exists {
		utils.RespondWithError(c, http.StatusInternalServerError, "failed to get role")
		return
	}
	userRole := roleValue.(string)
	if userRole != "HR" && userRole != "SUPERADMIN" && userRole != "ADMIN" {
		utils.RespondWithError(c, http.StatusUnauthorized, "not permitted to assign manager")
		return
	}

	// 2. Target employee ID
	empIDParam := c.Param("id")
	empID, err := uuid.Parse(empIDParam)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "invalid employee ID")
		return
	}

	// 3. Bind input
	var input UpdateManagerInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "invalid JSON")
		return
	}

	managerID, err := uuid.Parse(input.ManagerID)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "invalid manager ID")
		return
	}

	// 4. Prevent assigning self
	if empID == managerID {
		utils.RespondWithError(c, http.StatusBadRequest, "employee cannot be their own manager")
		return
	}

	// 5. Check if manager exists
	var existsManager bool
	err = s.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM TBL_EMPLOYEE WHERE ID=$1)", managerID).Scan(&existsManager)
	if err != nil || !existsManager {
		utils.RespondWithError(c, http.StatusNotFound, "manager not found")
		return
	}

	// 6. Update manager
	_, err = s.DB.Exec(`
        UPDATE TBL_EMPLOYEE
        SET MANAGER_ID=$1, UPDATED_AT=NOW()
        WHERE ID=$2
    `, managerID, empID)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "failed to update manager")
		return
	}

	// 7. Response
	c.JSON(http.StatusOK, gin.H{
		"message":     "manager updated successfully",
		"employee_id": empID,
		"manager_id":  managerID,
	})
}

// GetEmployeeReports - GET /api/employees/:id/reports
func (s *HandlerFunc) GetEmployeeReports(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Get employee reports"})
}
