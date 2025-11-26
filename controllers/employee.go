package controllers

import (
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
	role, _ := c.Get("role")
	r := role.(string)

	if r != "SUPERADMIN" && r != "ADMIN" {
		utils.RespondWithError(c, http.StatusUnauthorized, "not permitted")
		return
	}

	employees, err := h.Query.GetAllEmployees()
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(200, gin.H{
		"message":   "Employees fetched",
		"employees": employees,
	})
}

func (h *HandlerFunc) GetEmployeeById(c *gin.Context) {

}

func (h *HandlerFunc) CreateEmployee(c *gin.Context) {
	role := c.GetString("role")
	if role != "SUPERADMIN" && role != "Admin" {
		utils.RespondWithError(c, http.StatusUnauthorized, "not permitted")
		return
	}

	var input models.EmployeeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	if !strings.HasSuffix(input.Email, "@zenithive.com") {
		utils.RespondWithError(c, 400, "email must end with @zenithive.com")
		return
	}

	// EMAIL EXIST CHECK
	exists, err := h.Query.CheckEmailExists(input.Email)
	if err != nil {
		utils.RespondWithError(c, 500, err.Error())
		return
	}
	if exists {
		utils.RespondWithError(c, 400, "email already exists")
		return
	}

	// GET ROLE ID
	roleID, err := h.Query.GetRoleID(input.Role)
	if err != nil {
		utils.RespondWithError(c, 400, "role not found")
		return
	}

	// HASH PASSWORD
	hash, _ := utils.HashPassword(input.Password)

	// INSERT
	err = h.Query.InsertEmployee(
		input.FullName, input.Email,
		roleID, hash,
		input.Salary, input.JoiningDate,
	)
	if err != nil {
		utils.RespondWithError(c, 500, "failed to create employee")
		return
	}

	c.JSON(201, gin.H{"message": "employee created"})
}
func (h *HandlerFunc) UpdateEmployeeRole(c *gin.Context) {
	// ---------------------------
	// 1️ Check permission
	// ---------------------------
	role := c.GetString("role")
	if role != "SUPERADMIN" && role != "ADMIN" {
		utils.RespondWithError(c, 401, "not permitted")
		return
	}

	// ---------------------------
	// 2️ Parse Employee ID
	// ---------------------------
	empIDStr := c.Param("id")
	empID, err := uuid.Parse(empIDStr)
	if err != nil {
		utils.RespondWithError(c, 400, "invalid employee ID")
		return
	}

	// ---------------------------
	// 3️ Bind input JSON
	// ---------------------------
	var input UpdateRoleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, 400, "invalid input: "+err.Error())
		return
	}

	// ---------------------------
	// 4️ Fetch current role & status
	// ---------------------------
	currentRole, isManager, err := h.Query.GetEmployeeCurrentRoleAndManagerStatus(empID)
	if err != nil {
		utils.RespondWithError(c, 500, "failed to fetch employee role: "+err.Error())
		return
	}

	// ---------------------------
	// 5️ Check if role is unchanged
	// ---------------------------
	if currentRole == input.Role {
		utils.RespondWithError(c, 400, "employee already has this role")
		return
	}

	// ---------------------------
	// 6️ Edge Case: Employee is a Manager
	// ---------------------------
	if isManager && input.Role != "MANAGER" {
		// Employee manages other employees → cannot remove MANAGER role
		utils.RespondWithError(c, 403, "cannot change role of employee who is a manager with subordinates")
		return
	}

	// ---------------------------
	// 7️ Update Role
	// ---------------------------
	updatedID, err := h.Query.UpdateEmployeeRole(empID, input.Role)
	if err != nil {
		utils.RespondWithError(c, 500, "failed to update role: "+err.Error())
		return
	}

	// ---------------------------
	// 8️⃣ Response
	// ---------------------------
	c.JSON(200, gin.H{
		"message":     "role updated successfully",
		"employee_id": updatedID,
		"old_role":    currentRole,
		"new_role":    input.Role,
	})
}

func (h *HandlerFunc) DeleteEmployeeStatus(c *gin.Context) {

	// Read ID
	idParam := c.Param("id")
	empID, err := uuid.Parse(idParam)
	if err != nil {
		utils.RespondWithError(c, 400, "invalid employee id")
		return
	}

	// Role check
	role, _ := c.Get("role")
	r := role.(string)

	if r != "SUPERADMIN" && r != "ADMIN" {
		utils.RespondWithError(c, 401, "not permitted")
		return
	}

	// Toggle using same method name
	newStatus, err := h.Query.DeleteEmployeeStatus(empID)
	if err != nil {
		utils.RespondWithError(c, 500, err.Error())
		return
	}

	c.JSON(200, gin.H{
		"message":    "Employee status updated successfully",
		"new_status": newStatus,
	})
}
func (h *HandlerFunc) UpdateEmployeeManager(c *gin.Context) {
	// 1️ Permission check
	role := c.GetString("role")
	if role != "SUPERADMIN" && role != "ADMIN" && role != "HR" {
		utils.RespondWithError(c, 401, "not permitted")
		return
	}

	// 2️⃣ Parse Employee ID
	empID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondWithError(c, 400, "invalid employee ID")
		return
	}

	// 3️⃣ Parse Manager ID
	var input UpdateManagerInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, 400, "invalid input: "+err.Error())
		return
	}
	managerID, err := uuid.Parse(input.ManagerID)
	if err != nil {
		utils.RespondWithError(c, 400, "invalid manager ID")
		return
	}

	// 4️⃣ Self assignment check
	if empID == managerID {
		utils.RespondWithError(c, 400, "cannot assign employee as their own manager")
		return
	}

	// 5️⃣ Check if employee already has a manager
	var existingManager uuid.UUID
	err = h.Query.DB.Get(&existingManager, "SELECT manager_id FROM Tbl_Employee WHERE id=$1", empID)
	if err != nil {
		utils.RespondWithError(c, 404, "employee not found")
		return
	}

	if existingManager != uuid.Nil {
		utils.RespondWithError(c, 400, "employee already has a manager assigned")
		return
	}

	// 6️⃣ Validate Manager exists, active and role = MANAGER
	var mgrRole, mgrStatus string
	err = h.Query.DB.Get(&mgrRole, "SELECT r.type FROM Tbl_Employee e JOIN Tbl_Role r ON e.role_id = r.id WHERE e.id=$1", managerID)
	if err != nil {
		utils.RespondWithError(c, 404, "manager not found")
		return
	}
	err = h.Query.DB.Get(&mgrStatus, "SELECT status FROM Tbl_Employee WHERE id=$1", managerID)
	if err != nil || mgrStatus != "active" {
		utils.RespondWithError(c, 403, "manager is deactivated")
		return
	}
	if mgrRole != "MANAGER" {
		utils.RespondWithError(c, 400, "assigned employee is not a manager")
		return
	}

	// 7️⃣ Update manager
	err = h.Query.UpdateManager(empID, managerID)
	if err != nil {
		utils.RespondWithError(c, 500, "failed to update manager: "+err.Error())
		return
	}

	// 8️⃣ Success response
	c.JSON(200, gin.H{
		"message":     "manager updated successfully",
		"employee_id": empID,
		"manager_id":  managerID,
	})
}

// func (h *HandlerFunc) UpdateEmployeeInfo(c *gin.Context) {
// 	c.JSON(http.StatusOK, gin.H{"message": "Employee info updated"})
// }

// GetEmployeeReports - GET /api/employees/:id/reports
func (s *HandlerFunc) GetEmployeeReports(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Get employee reports"})
}
