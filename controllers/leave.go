package controllers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sanjayk-eng/UserMenagmentSystem_Backend/models"
	"github.com/sanjayk-eng/UserMenagmentSystem_Backend/service"
	"github.com/sanjayk-eng/UserMenagmentSystem_Backend/utils"
	"github.com/sanjayk-eng/UserMenagmentSystem_Backend/utils/common"
	"github.com/sanjayk-eng/UserMenagmentSystem_Backend/utils/constant"
)

type Leave struct {
	ID           uuid.UUID  `db:"id"`
	EmployeeID   uuid.UUID  `db:"employee_id"`
	LeaveTypeID  int        `db:"leave_type_id"`
	StartDate    time.Time  `db:"start_date"`
	EndDate      time.Time  `db:"end_date"`
	Days         float64    `db:"days"`
	Status       string     `db:"status"`
	AppliedByID  *uuid.UUID `db:"applied_by"`
	ApprovedByID *uuid.UUID `db:"approved_by"`
	Reason       string     `db:"reason"` //  ADD THIS
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
}

// ApplyLeave - POST /api/leaves/apply
// func (h *HandlerFunc) ApplyLeave(c *gin.Context) {
// 	// Extract Employee Info from Middleware
// 	empIDRaw, ok := c.Get("user_id")
// 	if !ok {
// 		utils.RespondWithError(c, http.StatusUnauthorized, "Employee ID missing")
// 		return
// 	}

// 	empIDStr, ok := empIDRaw.(string)
// 	if !ok {
// 		utils.RespondWithError(c, http.StatusInternalServerError, "Invalid employee ID format")
// 		return
// 	}

// 	employeeID, err := uuid.Parse(empIDStr)
// 	if err != nil {
// 		utils.RespondWithError(c, http.StatusInternalServerError, "Invalid employee UUID")
// 		return
// 	}

// 	//role, ok := c.Get("role")
// 	// if !ok  {
// 	// 	utils.RespondWithError(c, http.StatusForbidden, "Only employees can apply leave")
// 	// 	return
// 	// }

// 	// Validate Employee Status
// 	var empStatus string
// 	err = h.Query.DB.Get(&empStatus, `
// 		SELECT status FROM Tbl_Employee WHERE id=$1
// 	`, employeeID)

// 	if err != nil {
// 		utils.RespondWithError(c, 500, "Failed to verify employee status")
// 		return
// 	}

// 	if empStatus == "deactive" {
// 		utils.RespondWithError(c, 403, "Your account is deactivated. You cannot apply leave")
// 		return
// 	}

// 	// Bind Input JSON
// 	var input models.LeaveInput
// 	if err := c.ShouldBindJSON(&input); err != nil {
// 		utils.RespondWithError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
// 		return
// 	}

// 	input.EmployeeID = employeeID

// 	// Validate reason - Enhanced validation
// 	if input.Reason == "" {
// 		utils.RespondWithError(c, 400, "Leave reason is required. Please provide a valid reason for your leave request")
// 		return
// 	}

// 	// Trim whitespace and check minimum length
// 	input.Reason = strings.TrimSpace(input.Reason)
// 	if len(input.Reason) < 10 {
// 		utils.RespondWithError(c, 400, "Leave reason must be at least 10 characters long. Please provide a detailed reason")
// 		return
// 	}

// 	if len(input.Reason) > 500 {
// 		utils.RespondWithError(c, 400, "Leave reason is too long. Maximum 500 characters allowed")
// 		return
// 	}

// 	// Validate Date
// 	today := time.Now().Truncate(24 * time.Hour)
// 	if input.StartDate.Before(today) {
// 		utils.RespondWithError(c, 400, "Start date cannot be earlier than today")
// 		return
// 	}

// 	if input.EndDate.Before(input.StartDate) {
// 		utils.RespondWithError(c, 400, "End date cannot be earlier than start date")
// 		return
// 	}

// 	// Start Transaction
// 	tx, err := h.Query.DB.Beginx()
// 	if err != nil {
// 		utils.RespondWithError(c, 500, "Failed to start transaction")
// 		return
// 	}
// 	defer func() {
// 		if err != nil {
// 			tx.Rollback()
// 		}
// 	}()

// 	// Calculate Working Days
// 	leaveDays, err := service.CalculateWorkingDays(tx, input.StartDate, input.EndDate)
// 	if err != nil {
// 		utils.RespondWithError(c, 500, "Failed to calculate leave days: "+err.Error())
// 		return
// 	}
// 	if leaveDays <= 0 {
// 		utils.RespondWithError(c, 400, "Leave days must be greater than 0")
// 		return
// 	}
// 	input.Days = &leaveDays

// 	// Validate Leave Type
// 	var leaveType struct {
// 		DefaultEntitlement int `db:"default_entitlement"`
// 	}
// 	err = tx.Get(&leaveType,
// 		"SELECT default_entitlement FROM Tbl_Leave_type WHERE id=$1",
// 		input.LeaveTypeID,
// 	)
// 	if err == sql.ErrNoRows {
// 		utils.RespondWithError(c, 400, "Invalid leave type")
// 		return
// 	}
// 	if err != nil {
// 		utils.RespondWithError(c, 500, "Failed to fetch leave type")
// 		return
// 	}

// 	// Get/Create Leave Balance
// 	var balance float64
// 	err = tx.Get(&balance, `
// 		SELECT closing
// 		FROM Tbl_Leave_balance
// 		WHERE employee_id=$1 AND leave_type_id=$2
// 		AND year = EXTRACT(YEAR FROM CURRENT_DATE)
// 	`, employeeID, input.LeaveTypeID)

// 	if err == sql.ErrNoRows {
// 		balance = float64(leaveType.DefaultEntitlement)

// 		_, err = tx.Exec(`
// 			INSERT INTO Tbl_Leave_balance
// 				(employee_id, leave_type_id, year, opening, accrued, used, adjusted, closing)
// 			VALUES ($1, $2, EXTRACT(YEAR FROM CURRENT_DATE), $3, 0, 0, 0, $3)
// 		`, employeeID, input.LeaveTypeID, leaveType.DefaultEntitlement)
// 		if err != nil {
// 			utils.RespondWithError(c, 500, "Failed to create leave balance")
// 			return
// 		}
// 	} else if err != nil {
// 		utils.RespondWithError(c, 500, "Failed to fetch leave balance")
// 		return
// 	}

// 	// Check Enough Leaves
// 	if balance < leaveDays {
// 		utils.RespondWithError(c, 400, "Insufficient leave balance")
// 		return
// 	}

// 	// Overlapping Leave Check - Only check Pending and Approved leaves
// 	var overlappingLeaves []struct {
// 		ID        uuid.UUID `db:"id"`
// 		LeaveType string    `db:"leave_type"`
// 		StartDate time.Time `db:"start_date"`
// 		EndDate   time.Time `db:"end_date"`
// 		Status    string    `db:"status"`
// 	}

// 	err = tx.Select(&overlappingLeaves, `
// 		SELECT l.id, lt.name as leave_type, l.start_date, l.end_date, l.status
// 		FROM Tbl_Leave l
// 		JOIN Tbl_Leave_type lt ON l.leave_type_id = lt.id
// 		WHERE l.employee_id=$1
// 		AND l.status IN ('Pending','APPROVED')
// 		AND l.start_date <= $2
// 		AND l.end_date >= $3
// 	`, employeeID, input.EndDate, input.StartDate)

// 	if err != nil {
// 		utils.RespondWithError(c, 500, "Failed to check overlapping leave")
// 		return
// 	}

// 	if len(overlappingLeaves) > 0 {
// 		// Build detailed error message
// 		overlap := overlappingLeaves[0]
// 		utils.RespondWithError(c, 400,
// 			fmt.Sprintf("Overlapping leave exists: %s from %s to %s (Status: %s). Please cancel or modify the existing leave first",
// 				overlap.LeaveType,
// 				overlap.StartDate.Format("2006-01-02"),
// 				overlap.EndDate.Format("2006-01-02"),
// 				overlap.Status))
// 		return
// 	}

// 	// INSERT Leave with Reason
// 	var leaveID uuid.UUID
// 	err = tx.QueryRow(`
// 		INSERT INTO Tbl_Leave
// 		(employee_id, leave_type_id, start_date, end_date, days, status, reason)
// 		VALUES ($1,$2,$3,$4,$5,'Pending',$6)
// 		RETURNING id
// 	`,
// 		employeeID,
// 		input.LeaveTypeID,
// 		input.StartDate,
// 		input.EndDate,
// 		leaveDays,
// 		input.Reason,
// 	).Scan(&leaveID)

// 	if err != nil {
// 		utils.RespondWithError(c, 500, "Failed to apply leave: "+err.Error())
// 		return
// 	}

// 	// Commit Transaction
// 	err = tx.Commit()
// 	if err != nil {
// 		utils.RespondWithError(c, 500, "Failed to commit transaction")
// 		return
// 	}

// 	// Send notification to manager/admin (async)
// 	go func() {
// 		// Get employee details

// 		// Get leave type name
// 		var leaveTypeName string
// 		h.Query.DB.Get(&leaveTypeName, "SELECT name FROM Tbl_Leave_type WHERE id=$1", input.LeaveTypeID)

// 		recipients, err := h.Query.GetAdminAndEmployeeEmail(employeeID)
// 		if err != nil {
// 			fmt.Printf("Failed to get notification recipients , email: %v\n", err)
// 		}

// 		empDetails, err := h.Query.GetEmployeeDetailsForNotification(employeeID)
// 		if err != nil {
// 			fmt.Printf("Failed to get employee details for notification: %v\n", err)
// 		}
// 		// Send notification to all recipients
// 		if len(recipients) > 0 {
// 			err := utils.SendLeaveApplicationEmail(
// 				recipients,
// 				empDetails.FullName,
// 				leaveTypeName,
// 				input.StartDate.Format("2006-01-02"),
// 				input.EndDate.Format("2006-01-02"),
// 				leaveDays,
// 				input.Reason,
// 			)
// 			if err != nil {
// 				fmt.Printf("Failed to send leave application email: %v\n", err)
// 			}
// 		}
// 	}()

// 	// Response
// 	c.JSON(200, gin.H{
// 		"message":  "Leave applied successfully",
// 		"leave_id": leaveID,
// 		"days":     leaveDays,
// 		"reason":   input.Reason,
// 	})
// }

// AdminAddLeave - POST /api/leaves/admin-add

func (h *HandlerFunc) ApplyLeave(c *gin.Context) {

	// Extract Employee Info
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

	employeeID, err := uuid.Parse(empIDStr)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Invalid employee UUID")
		return
	}

	// Validate Employee Status
	empStatus, err := h.Query.GetEmployeeStatus(employeeID)
	if err != nil {
		utils.RespondWithError(c, 500, "Failed to verify employee status")
		return
	}
	if empStatus == "deactive" {
		utils.RespondWithError(c, 403, "Your account is deactivated. You cannot apply leave")
		return
	}

	// Bind Input
	var input models.LeaveInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}
	input.EmployeeID = employeeID

	// Validate Reason
	input.Reason = strings.TrimSpace(input.Reason)
	if len(input.Reason) < 10 {
		utils.RespondWithError(c, 400, "Leave reason must be at least 10 characters long")
		return
	}
	if len(input.Reason) > 500 {
		utils.RespondWithError(c, 400, "Leave reason is too long. Maximum 500 characters allowed")
		return
	}

	// Validate Dates
	today := time.Now().Truncate(24 * time.Hour)
	if input.StartDate.Before(today) {
		utils.RespondWithError(c, 400, "Start date cannot be earlier than today")
		return
	}
	if input.EndDate.Before(input.StartDate) {
		utils.RespondWithError(c, 400, "End date cannot be earlier than start date")
		return
	}

	// Final Leave ID to return
	var leaveID uuid.UUID
	var Days float64

	// Execute Transaction
	err = common.ExecuteTransaction(c, h.Query.DB, func(tx *sqlx.Tx) error {

		// Working days
		leaveDays, err := service.CalculateWorkingDays(tx, input.StartDate, input.EndDate)
		if err != nil {
			return utils.CustomErr(c, 500, "Failed to calculate leave days: "+err.Error())
		}
		if leaveDays <= 0 {
			return utils.CustomErr(c, 400, "Leave days must be greater than 0")
		}
		input.Days = &leaveDays
		Days = leaveDays

		// Leave Type
		leaveType, err := h.Query.GetLeaveTypeByIdTx(tx, input.LeaveTypeID)
		if err == sql.ErrNoRows {
			return utils.CustomErr(c, 400, "Invalid leave type")
		}
		if err != nil {
			return utils.CustomErr(c, 500, "Failed to fetch leave type: "+err.Error())
		}

		// Leave Balance
		balance, err := h.Query.GetLeaveBalance(tx, employeeID, input.LeaveTypeID)
		if err == sql.ErrNoRows {
			balance = float64(leaveType.DefaultEntitlement)
			if err := h.Query.CreateLeaveBalance(tx, employeeID, input.LeaveTypeID, leaveType.DefaultEntitlement); err != nil {
				return utils.CustomErr(c, 500, "Failed to create leave balance: "+err.Error())
			}
		} else if err != nil {
			return utils.CustomErr(c, 500, "Failed to fetch leave balance: "+err.Error())
		}

		// Check balance
		if balance < leaveDays {
			return utils.CustomErr(c, 400, "Insufficient leave balance")
		}

		// Overlapping Leave
		overlaps, err := h.Query.GetOverlappingLeaves(tx, employeeID, input.StartDate, input.EndDate)
		if err != nil {
			return utils.CustomErr(c, 500, "Failed to check overlapping leave")
		}
		if len(overlaps) > 0 {
			ov := overlaps[0]
			return utils.CustomErr(c, 400, fmt.Sprintf(
				"Overlapping leave exists: %s from %s to %s (Status: %s). Please cancel or modify the existing leave first",
				ov.LeaveType,
				ov.StartDate.Format("2006-01-02"),
				ov.EndDate.Format("2006-01-02"),
				ov.Status,
			))
		}

		// Insert Leave
		id, err := h.Query.InsertLeave(tx, employeeID, input.LeaveTypeID, input.StartDate, input.EndDate, leaveDays, input.Reason)
		if err != nil {
			return utils.CustomErr(c, 500, "Failed to apply leave: "+err.Error())
		}
		leaveID = id

		// Log Entry
		data := &utils.Common{
			Component:  constant.ComponentLeave,
			Action:     constant.ActionCreate,
			FromUserID: employeeID,
		}
		if err := common.AddLog(data, tx); err != nil {
			return utils.CustomErr(c, 500, "Failed to create leave log: "+err.Error())
		}

		return nil // IMPORTANT FIX
	})

	// If transaction returned an error, stop (CustomErr already responded)
	if err != nil {
		utils.RespondWithError(c, 500, "Failed to update settings: "+err.Error())
		return
	}

	go func() {
		leaveType, _ := h.Query.GetLeaveTypeById(input.LeaveTypeID)

		recipients, err := h.Query.GetAdminAndEmployeeEmail(employeeID)
		if err != nil {
			fmt.Printf("Failed to get notification recipients: %v\n", err)
			return
		}

		empDetails, err := h.Query.GetEmployeeDetailsForNotification(employeeID)
		if err != nil {
			fmt.Printf("Failed to get employee details for notification: %v\n", err)
			return
		}

		if len(recipients) > 0 {
			utils.SendLeaveApplicationEmail(
				recipients,
				empDetails.FullName,
				leaveType.Name,
				input.StartDate.Format("2006-01-02"),
				input.EndDate.Format("2006-01-02"),
				*input.Days,
				input.Reason,
			)
		}
	}()

	// Send response
	c.JSON(200, gin.H{
		"message":  "Leave applied successfully",
		"leave_id": leaveID,
		"days":     Days,
		"reason":   input.Reason,
	})
}

func (s *HandlerFunc) AdminAddLeavePolicy(c *gin.Context) {
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
	err := s.Query.DB.QueryRow(query, input.Name, *input.IsPaid, *input.DefaultEntitlement).
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
	leaveType, err := s.Query.GetAllLeaveType()
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch leave types: "+err.Error())
		return
	}
	c.JSON(http.StatusOK, leaveType) // send models directly
}

// ActionLeave - POST /api/leaves/:id/action
// Two-level approval/rejection system:
// APPROVAL FLOW:
// 1. MANAGER approves → Status: MANAGER_APPROVED (no balance deduction)
// 2. ADMIN/SUPERADMIN finalizes → Status: APPROVED (balance deducted)
// REJECTION FLOW:
// 1. MANAGER rejects → Status: MANAGER_REJECTED (pending final rejection)
// 2. ADMIN/SUPERADMIN finalizes → Status: REJECTED (final rejection)
func (s *HandlerFunc) ActionLeave(c *gin.Context) {
	roleRaw, _ := c.Get("role")
	role := roleRaw.(string)

	if role == "EMPLOYEE" {
		utils.RespondWithError(c, 403, "Employees cannot approve leaves")
		return
	}

	approverIDRaw, _ := c.Get("user_id")
	approverID, _ := uuid.Parse(approverIDRaw.(string))

	leaveID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondWithError(c, 400, "Invalid leave ID")
		return
	}

	var body struct {
		Action string `json:"action" validate:"required"` // APPROVE/REJECT
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.RespondWithError(c, 400, "Invalid payload: "+err.Error())
		return
	}

	body.Action = strings.ToUpper(body.Action)
	if body.Action != "APPROVE" && body.Action != "REJECT" {
		utils.RespondWithError(c, 400, "Action must be APPROVE or REJECT")
		return
	}

	tx, err := s.Query.DB.Beginx()
	if err != nil {
		utils.RespondWithError(c, 500, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	var leave Leave
	err = tx.Get(&leave, `SELECT * FROM Tbl_Leave WHERE id=$1 FOR UPDATE`, leaveID)
	if err != nil {
		utils.RespondWithError(c, 404, "Leave not found: "+err.Error())
		return
	}

	//  Prevent self-approval
	if leave.EmployeeID == approverID {
		utils.RespondWithError(c, 403, "You cannot approve your own leave request")
		return
	}

	//  MANAGER validation
	if role == "MANAGER" {
		// Check manager permission setting
		exists, err := s.Query.ChackManagerPermission()
		if err != nil {
			utils.RespondWithError(c, 500, "Failed to get manager permission")
			return
		}
		if !exists {
			utils.RespondWithError(c, 403, "Manager approval is not enabled")
			return
		}

		// Verify reporting relationship
		var managerID uuid.UUID
		err = tx.Get(&managerID, "SELECT manager_id FROM Tbl_Employee WHERE id=$1", leave.EmployeeID)
		if err != nil {
			utils.RespondWithError(c, 500, "Failed to verify reporting relationship")
			return
		}

		if managerID != approverID {
			utils.RespondWithError(c, 403, "You can only approve leaves of employees who report to you")
			return
		}

		// Manager can only act on Pending leaves
		if leave.Status != "Pending" {
			utils.RespondWithError(c, 400, fmt.Sprintf("Cannot process leave with status: %s", leave.Status))
			return
		}
	}

	// 🔒 ADMIN/SUPERADMIN validation
	if role == "ADMIN" || role == "SUPERADMIN" {
		// Admin can act on Pending, MANAGER_APPROVED, or MANAGER_REJECTED leaves
		if leave.Status != "Pending" && leave.Status != "MANAGER_APPROVED" && leave.Status != "MANAGER_REJECTED" {
			utils.RespondWithError(c, 400, fmt.Sprintf("Cannot process leave with status: %s", leave.Status))
			return
		}
	}

	// ========================================
	// REJECT ACTION - Two-Step Process
	// ========================================
	if body.Action == "REJECT" {
		// MANAGER REJECTION (First Level)
		if role == "MANAGER" {
			_, err = tx.Exec(`UPDATE Tbl_Leave SET status='MANAGER_REJECTED', approved_by=$2, updated_at=NOW() WHERE id=$1`, leaveID, approverID)
			if err != nil {
				utils.RespondWithError(c, 500, "Failed to reject leave: "+err.Error())
				return
			}

			empDetails, err := s.Query.GetEmployeeDetailsForNotification(leave.EmployeeID)
			if err != nil {
				fmt.Printf("Failed to get employee details for notification: %v\n", err)
			}

			var leaveTypeName string
			s.Query.DB.Get(&leaveTypeName, "SELECT name FROM Tbl_Leave_type WHERE id=$1", leave.LeaveTypeID)

			// Fetch approver's full name
			var approverName string
			s.Query.DB.Get(&approverName, "SELECT full_name FROM Tbl_Employee WHERE id=$1", approverID)

			recipients, err := s.Query.GetAdminAndEmployeeEmail(leave.EmployeeID)
			if err != nil {
				fmt.Printf("Failed to get admin and employee emails for notification: %v\n", err)
			}
			recipients = append(recipients)

			tx.Commit()

			// Send final rejection notification
			if len(recipients) > 0 {
				go func() {
					utils.SendLeaveManagerRejectionEmail(
						recipients,
						empDetails.Email,
						empDetails.FullName,
						leaveTypeName,
						leave.StartDate.Format("2006-01-02"),
						leave.EndDate.Format("2006-01-02"),
						leave.Days,
						approverName,
					)
				}()
			}

			c.JSON(200, gin.H{
				"message": "Leave rejected by manager. Pending final rejection from ADMIN/SUPERADMIN",
				"status":  "MANAGER_REJECTED",
			})
			return
		}

		// ADMIN/SUPERADMIN FINAL REJECTION (Second Level)
		if role == "ADMIN" || role == "SUPERADMIN" {
			_, err = tx.Exec(`UPDATE Tbl_Leave SET status='REJECTED', approved_by=$2, updated_at=NOW() WHERE id=$1`, leaveID, approverID)
			if err != nil {
				utils.RespondWithError(c, 500, "Failed to finalize leave rejection: "+err.Error())
				return
			}

			// Fetch employee details
			empDetails, err := s.Query.GetEmployeeDetailsForNotification(leave.EmployeeID)
			if err != nil {
				fmt.Printf("Failed to get employee details for notification: %v\n", err)
			}

			var leaveTypeName string
			s.Query.DB.Get(&leaveTypeName, "SELECT name FROM Tbl_Leave_type WHERE id=$1", leave.LeaveTypeID)

			// Fetch approver's full name
			var approverName string
			s.Query.DB.Get(&approverName, "SELECT full_name FROM Tbl_Employee WHERE id=$1", approverID)

			recipients, err := s.Query.GetAdminAndEmployeeEmail(leave.EmployeeID)
			if err != nil {
				fmt.Printf("Failed to get admin and employee emails for notification: %v\n", err)
			}
			recipients = append(recipients)

			tx.Commit()

			// Send final rejection notification
			if len(recipients) > 0 {
				go func() {
					utils.SendLeaveRejectionEmail(
						recipients,
						empDetails.Email,
						empDetails.FullName,
						leaveTypeName,
						leave.StartDate.Format("2006-01-02"),
						leave.EndDate.Format("2006-01-02"),
						leave.Days,
						approverName,
					)
				}()
			}

			c.JSON(200, gin.H{
				"message": "Leave finalized and rejected successfully",
				"status":  "REJECTED",
			})
			return
		}

		// Should not reach here
		utils.RespondWithError(c, 500, "Unexpected error in leave rejection process")
		return
	}

	// ========================================
	// APPROVE ACTION
	// ========================================

	// Check balance before any approval
	var currentBalance float64
	err = tx.Get(&currentBalance, `
		SELECT closing 
		FROM Tbl_Leave_balance 
		WHERE employee_id=$1 AND leave_type_id=$2 
		AND year = EXTRACT(YEAR FROM CURRENT_DATE)
	`, leave.EmployeeID, leave.LeaveTypeID)

	if err != nil {
		utils.RespondWithError(c, 500, "Failed to fetch leave balance: "+err.Error())
		return
	}

	if currentBalance < leave.Days {
		utils.RespondWithError(c, 400, fmt.Sprintf("Cannot approve: Insufficient leave balance. Available: %.1f days, Required: %.1f days", currentBalance, leave.Days))
		return
	}

	// MANAGER APPROVAL (First Level)
	if role == "MANAGER" {
		_, err = tx.Exec(`UPDATE Tbl_Leave SET status='MANAGER_APPROVED', approved_by=$2, updated_at=NOW() WHERE id=$1`, leaveID, approverID)
		if err != nil {
			utils.RespondWithError(c, 500, "Failed to approve leave: "+err.Error())
			return
		}

		empDetails, err := s.Query.GetEmployeeDetailsForNotification(leave.EmployeeID)
		if err != nil {
			fmt.Printf("Failed to get employee details for notification: %v\n", err)
		}

		var leaveTypeName string
		s.Query.DB.Get(&leaveTypeName, "SELECT name FROM Tbl_Leave_type WHERE id=$1", leave.LeaveTypeID)

		// Fetch approver's full name
		var approverName string
		s.Query.DB.Get(&approverName, "SELECT full_name FROM Tbl_Employee WHERE id=$1", approverID)

		tx.Commit()

		admins, err := s.Query.GetAdminAndEmployeeEmail(leave.EmployeeID)
		if err != nil {
			fmt.Printf("Failed to get admin and employee emails for notification: %v\n", err)
		}

		// Send final approval notification
		if empDetails.Email != "" {
			go func() {
				utils.SendLeaveManagerApprovalEmail(
					admins,
					empDetails.Email,
					empDetails.FullName,
					leaveTypeName,
					leave.StartDate.Format("2006-01-02"),
					leave.EndDate.Format("2006-01-02"),
					leave.Days,
					approverName,
				)
			}()
		}

		c.JSON(200, gin.H{
			"message": "Leave approved by manager. Pending final approval from ADMIN/SUPERADMIN",
			"status":  "MANAGER_APPROVED",
		})
		return
	}

	// ADMIN/SUPERADMIN FINAL APPROVAL (Second Level)
	if role == "ADMIN" || role == "SUPERADMIN" {
		// Update status to APPROVED
		_, err = tx.Exec(`UPDATE Tbl_Leave SET status='APPROVED', approved_by=$2, updated_at=NOW() WHERE id=$1`, leaveID, approverID)
		if err != nil {
			utils.RespondWithError(c, 500, "Failed to finalize leave approval: "+err.Error())
			return
		}

		// Deduct from leave balance
		_, err = tx.Exec(`UPDATE Tbl_Leave_balance SET used = used + $3, closing = closing - $3, updated_at = NOW() WHERE employee_id=$1 AND leave_type_id=$2`,
			leave.EmployeeID, leave.LeaveTypeID, leave.Days)
		if err != nil {
			utils.RespondWithError(c, 500, "Failed to update leave balance: "+err.Error())
			return
		}

		// Fetch employee details
		empDetails, err := s.Query.GetEmployeeDetailsForNotification(leave.EmployeeID)
		if err != nil {
			fmt.Printf("Failed to get employee details for notification: %v\n", err)
		}

		var leaveTypeName string
		s.Query.DB.Get(&leaveTypeName, "SELECT name FROM Tbl_Leave_type WHERE id=$1", leave.LeaveTypeID)

		// Fetch approver's full name
		var approverName string
		s.Query.DB.Get(&approverName, "SELECT full_name FROM Tbl_Employee WHERE id=$1", approverID)

		tx.Commit()

		admins, err := s.Query.GetAdminAndEmployeeEmail(leave.EmployeeID)
		if err != nil {
			fmt.Printf("Failed to get admin and employee emails for notification: %v\n", err)
		}

		// Send final approval notification
		if empDetails.Email != "" {
			go func() {
				utils.SendLeaveFinalApprovalEmail(
					admins,
					empDetails.Email,
					empDetails.FullName,
					leaveTypeName,
					leave.StartDate.Format("2006-01-02"),
					leave.EndDate.Format("2006-01-02"),
					leave.Days,
					approverName,
				)
			}()
		}

		c.JSON(200, gin.H{
			"message": "Leave finalized and approved successfully. Balance deducted.",
			"status":  "APPROVED",
		})
		return
	}

	// Should not reach here
	utils.RespondWithError(c, 500, "Unexpected error in leave approval process")
}

///  logic for payment and other

// CalculateWorkingDays returns only valid working days (Mon–Fri)
// excluding holidays stored in Tbl_Holiday.
// Uses same DB transaction for safety (tx).
func CalculateWorkingDays(tx *sqlx.Tx, start, end time.Time) (float64, error) {
	// 1️ Validate date range
	if end.Before(start) {
		return 0, fmt.Errorf("end date cannot be before start date")
	}

	// Normalize dates to midnight UTC to avoid timezone issues
	start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)
	end = time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, time.UTC)

	// 2️ Fetch holidays within range
	var holidays []time.Time
	err := tx.Select(&holidays,
		`SELECT date FROM Tbl_Holiday 
         WHERE date BETWEEN $1 AND $2`,
		start, end)

	if err != nil {
		return 0, fmt.Errorf("failed to fetch holidays: %v", err)
	}

	// Convert slice to a map for O(1) lookup
	holidayMap := make(map[string]bool)
	for _, h := range holidays {
		holidayMap[h.Format("2006-01-02")] = true
	}

	// 3️ Count working days
	workingDays := 0
	var workingDaysList []string // For debugging

	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		dayStr := d.Format("2006-01-02")
		weekday := d.Weekday()

		// Skip Saturday and Sunday
		if weekday == time.Saturday || weekday == time.Sunday {
			fmt.Printf("DEBUG: Skipping weekend: %s (%s)\n", dayStr, weekday)
			continue
		}

		// Skip holidays
		if holidayMap[dayStr] {
			fmt.Printf("DEBUG: Skipping holiday: %s\n", dayStr)
			continue
		}

		workingDays++
		workingDaysList = append(workingDaysList, fmt.Sprintf("%s (%s)", dayStr, weekday))
	}

	fmt.Printf("DEBUG: Working days calculated: %d - Days: %v\n", workingDays, workingDaysList)
	return float64(workingDays), nil
}
func (h *HandlerFunc) GetAllLeaves(c *gin.Context) {
	// 1️⃣ Get Role & User ID with validation
	role := c.GetString("role")
	if role == "" {
		utils.RespondWithError(c, http.StatusUnauthorized, "Role not found in context")
		return
	}

	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		utils.RespondWithError(c, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid user ID format: "+err.Error())
		return
	}

	// 2️⃣ Execute query based on role
	var result []models.LeaveResponse

	switch role {
	case "EMPLOYEE":
		// Employees can only see their own leaves
		query := `
		SELECT 
			l.id,
			e.full_name AS employee,
			lt.name AS leave_type,
			l.start_date,
			l.end_date,
			l.days,
			COALESCE(l.reason, '') AS reason,
			l.status,
			l.created_at AS applied_at
		FROM Tbl_Leave l
		INNER JOIN Tbl_Employee e ON l.employee_id = e.id
		INNER JOIN Tbl_Leave_Type lt ON lt.id = l.leave_type_id
		WHERE l.employee_id = $1
		ORDER BY l.created_at DESC`

		err = h.Query.DB.Select(&result, query, userID)

	case "MANAGER":
		// Manager can see: their own leaves + their team members' leaves
		query := `
		SELECT 
			l.id,
			e.full_name AS employee,
			lt.name AS leave_type,
			l.start_date,
			l.end_date,
			l.days,
			COALESCE(l.reason, '') AS reason,
			l.status,
			l.created_at AS applied_at
		FROM Tbl_Leave l
		INNER JOIN Tbl_Employee e ON l.employee_id = e.id
		INNER JOIN Tbl_Leave_Type lt ON lt.id = l.leave_type_id
		WHERE (e.manager_id = $1 OR l.employee_id = $1)
		ORDER BY l.created_at DESC`

		err = h.Query.DB.Select(&result, query, userID)

	case "HR", "ADMIN", "SUPERADMIN":
		// HR, Admin and SuperAdmin can see all leaves
		query := `
		SELECT 
			l.id,
			e.full_name AS employee,
			lt.name AS leave_type,
			l.start_date,
			l.end_date,
			l.days,
			COALESCE(l.reason, '') AS reason,
			l.status,
			l.created_at AS applied_at
		FROM Tbl_Leave l
		INNER JOIN Tbl_Employee e ON l.employee_id = e.id
		INNER JOIN Tbl_Leave_Type lt ON lt.id = l.leave_type_id
		ORDER BY l.created_at DESC`

		err = h.Query.DB.Select(&result, query)

	default:
		utils.RespondWithError(c, http.StatusForbidden, "Invalid role: "+role)
		return
	}

	// 3️⃣ Handle query errors
	if err != nil {
		fmt.Printf("❌ GetAllLeaves DB Error: %v\n", err)
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch leaves: "+err.Error())
		return
	}

	// 4️⃣ Handle empty result
	if result == nil {
		result = []models.LeaveResponse{}
	}

	// 5️⃣ Return success with metadata
	c.JSON(http.StatusOK, gin.H{
		"message": "Leaves fetched successfully",
		"total":   len(result),
		"role":    role,
		"data":    result,
	})
}

// -------------------- Admin/Manager Add Leave --------------------
// AdminAddLeave - POST /api/leaves/admin-add
func (h *HandlerFunc) AdminAddLeave(c *gin.Context) {
	// ------------------------------
	// 1. Get Role & User ID from Middleware
	// ------------------------------
	roleValue, exists := c.Get("role")
	if !exists {
		utils.RespondWithError(c, http.StatusInternalServerError, "failed to get role")
		return
	}
	userRole := roleValue.(string)

	userIDRaw, _ := c.Get("user_id")
	currentUserID, _ := uuid.Parse(userIDRaw.(string))

	// ------------------------------
	// 2. Fetch company settings (for manager permission)
	// ------------------------------
	var settings struct {
		AllowManagerAddLeave bool `db:"allow_manager_add_leave"`
	}
	err := h.Query.DB.Get(&settings, "SELECT allow_manager_add_leave FROM Tbl_Company_Settings LIMIT 1")
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "failed to fetch company settings: "+err.Error())
		return
	}

	// ------------------------------
	// 3. Permission check
	// ------------------------------
	if userRole != "SUPERADMIN" &&
		userRole != "ADMIN" &&
		!(userRole == "MANAGER" && settings.AllowManagerAddLeave) {
		utils.RespondWithError(c, http.StatusUnauthorized, "not permitted to add leave")
		return
	}

	// ------------------------------
	// 4. Bind Input JSON
	// ------------------------------
	var input models.LeaveInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	if input.EmployeeID == uuid.Nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Employee ID is required")
		return
	}

	// Validate reason - Enhanced validation
	if input.Reason == "" {
		utils.RespondWithError(c, http.StatusBadRequest, "Leave reason is required. Please provide a valid reason for this leave")
		return
	}

	input.Reason = strings.TrimSpace(input.Reason)
	if len(input.Reason) < 10 {
		utils.RespondWithError(c, http.StatusBadRequest, "Leave reason must be at least 10 characters long. Please provide a detailed reason")
		return
	}

	if len(input.Reason) > 500 {
		utils.RespondWithError(c, http.StatusBadRequest, "Leave reason is too long. Maximum 500 characters allowed")
		return
	}

	// ------------------------------
	// 5. Manager can only add leave for their team (if not self)
	// ------------------------------
	if userRole == "MANAGER" && input.EmployeeID != currentUserID {
		var managerID uuid.UUID
		err := h.Query.DB.Get(&managerID, "SELECT manager_id FROM Tbl_Employee WHERE id=$1", input.EmployeeID)
		if err != nil {
			utils.RespondWithError(c, http.StatusBadRequest, "Employee not found")
			return
		}
		if managerID != currentUserID {
			utils.RespondWithError(c, http.StatusForbidden, "Managers can only add leave for their team members")
			return
		}
	}

	// ------------------------------
	// 6. Start transaction
	// ------------------------------
	tx, err := h.Query.DB.Beginx()
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to start transaction")
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// ------------------------------
	// 7. Calculate leave days (working days only)
	// ------------------------------
	leaveDays, err := CalculateWorkingDays(tx, input.StartDate, input.EndDate)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to calculate leave days: "+err.Error())
		return
	}
	if leaveDays <= 0 {
		utils.RespondWithError(c, http.StatusBadRequest, "Leave days must be greater than 0")
		return
	}
	input.Days = &leaveDays

	// ------------------------------
	// 8. Validate leave type
	// ------------------------------
	var leaveType struct {
		DefaultEntitlement int `db:"default_entitlement"`
	}
	err = tx.Get(&leaveType, "SELECT default_entitlement FROM Tbl_Leave_type WHERE id=$1", input.LeaveTypeID)
	if err == sql.ErrNoRows {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid leave type")
		return
	} else if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch leave type: "+err.Error())
		return
	}

	// ------------------------------
	// 9. Get/Create leave balance
	// ------------------------------
	var balance float64
	err = tx.Get(&balance, `
		SELECT closing 
		FROM Tbl_Leave_balance 
		WHERE employee_id=$1 AND leave_type_id=$2 
		AND year = EXTRACT(YEAR FROM CURRENT_DATE)
	`, input.EmployeeID, input.LeaveTypeID)

	if err == sql.ErrNoRows {
		// Create leave balance if missing
		balance = float64(leaveType.DefaultEntitlement)
		_, err = tx.Exec(`
			INSERT INTO Tbl_Leave_balance
			(employee_id, leave_type_id, year, opening, accrued, used, adjusted, closing)
			VALUES ($1, $2, EXTRACT(YEAR FROM CURRENT_DATE), $3, 0, 0, 0, $3)
		`, input.EmployeeID, input.LeaveTypeID, leaveType.DefaultEntitlement)
		if err != nil {
			utils.RespondWithError(c, http.StatusInternalServerError, "Failed to create leave balance: "+err.Error())
			return
		}
	} else if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch leave balance: "+err.Error())
		return
	}

	// ------------------------------
	// 9.5. Check sufficient balance
	// ------------------------------
	if balance < leaveDays {
		utils.RespondWithError(c, http.StatusBadRequest, fmt.Sprintf("Insufficient leave balance. Available: %.1f days, Requested: %.1f days", balance, leaveDays))
		return
	}

	// ------------------------------
	// 10. Insert leave (status APPROVED)
	// ------------------------------
	var leaveID uuid.UUID
	err = tx.QueryRow(`
		INSERT INTO Tbl_Leave 
		(employee_id, leave_type_id, start_date, end_date, days, status, reason, applied_by, approved_by, created_at)
		VALUES ($1, $2, $3, $4, $5, 'APPROVED', $6, $7, $7, NOW())
		RETURNING id
	`, input.EmployeeID, input.LeaveTypeID, input.StartDate, input.EndDate, leaveDays, input.Reason, currentUserID).
		Scan(&leaveID)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to insert leave: "+err.Error())
		return
	}

	// ------------------------------
	// 11. Update leave balance
	// ------------------------------
	_, err = tx.Exec(`
		UPDATE Tbl_Leave_balance 
		SET used = used + $1, closing = closing - $1, updated_at = NOW()
		WHERE employee_id=$2 AND leave_type_id=$3
	`, leaveDays, input.EmployeeID, input.LeaveTypeID)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to update leave balance: "+err.Error())
		return
	}

	// ------------------------------
	// 12. Commit transaction
	// ------------------------------
	err = tx.Commit()
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	// ------------------------------
	// 13. Send notification to employee
	// ------------------------------
	go func() {
		var empDetails struct {
			Email    string `db:"email"`
			FullName string `db:"full_name"`
		}
		h.Query.DB.Get(&empDetails, "SELECT email, full_name FROM Tbl_Employee WHERE id=$1", input.EmployeeID)

		var leaveTypeName string
		h.Query.DB.Get(&leaveTypeName, "SELECT name FROM Tbl_Leave_type WHERE id=$1", input.LeaveTypeID)

		var addedByName string
		h.Query.DB.Get(&addedByName, "SELECT full_name FROM Tbl_Employee WHERE id=$1", currentUserID)

		err := utils.SendLeaveAddedByAdminEmail(
			empDetails.Email,
			empDetails.FullName,
			leaveTypeName,
			input.StartDate.Format("2006-01-02"),
			input.EndDate.Format("2006-01-02"),
			leaveDays,
			addedByName,
			userRole,
		)
		if err != nil {
			fmt.Printf("Failed to send leave added notification: %v\n", err)
		}
	}()

	// ------------------------------
	// 14. Response
	// ------------------------------
	c.JSON(http.StatusOK, gin.H{
		"message":  "Leave added successfully",
		"leave_id": leaveID,
		"days":     leaveDays,
	})
}

// CancelLeave - DELETE /api/leaves/:id/cancel
// Allows employees to cancel their own pending leaves
func (h *HandlerFunc) CancelLeave(c *gin.Context) {
	// Get user info from middleware
	userIDRaw, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDRaw.(string))

	roleRaw, _ := c.Get("role")
	role := roleRaw.(string)

	// Parse leave ID from URL
	leaveID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondWithError(c, 400, "Invalid leave ID")
		return
	}

	// Start transaction
	tx, err := h.Query.DB.Beginx()
	if err != nil {
		utils.RespondWithError(c, 500, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	// Fetch leave details
	var leave Leave
	err = tx.Get(&leave, `SELECT * FROM Tbl_Leave WHERE id=$1 FOR UPDATE`, leaveID)
	if err != nil {
		utils.RespondWithError(c, 404, "Leave not found")
		return
	}

	// Permission check - employees can only cancel their own leaves
	if role == "EMPLOYEE" && leave.EmployeeID != userID {
		utils.RespondWithError(c, 403, "You can only cancel your own leave applications")
		return
	}

	// Check if leave can be cancelled
	if leave.Status == "APPROVED" {
		utils.RespondWithError(c, 400, "Cannot cancel approved leave. Please contact your manager or admin")
		return
	}

	if leave.Status == "REJECTED" {
		utils.RespondWithError(c, 400, "Leave is already rejected")
		return
	}

	if leave.Status == "CANCELLED" {
		utils.RespondWithError(c, 400, "Leave is already cancelled")
		return
	}

	// Update leave status to CANCELLED
	_, err = tx.Exec(`
		UPDATE Tbl_Leave 
		SET status='CANCELLED', updated_at=NOW() 
		WHERE id=$1
	`, leaveID)
	if err != nil {
		utils.RespondWithError(c, 500, "Failed to cancel leave: "+err.Error())
		return
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		utils.RespondWithError(c, 500, "Failed to commit transaction")
		return
	}

	c.JSON(200, gin.H{
		"message":  "Leave cancelled successfully",
		"leave_id": leaveID,
	})
}

// WithdrawLeave - POST /api/leaves/:id/withdraw
// Two-level withdrawal approval system:
// 1. MANAGER initiates withdrawal → Status: WITHDRAWAL_PENDING (no balance restoration)
// 2. ADMIN/SUPERADMIN finalizes → Status: WITHDRAWN (balance restored)
func (h *HandlerFunc) WithdrawLeave(c *gin.Context) {
	// 1️⃣ Get current user info
	role := c.GetString("role")
	currentUserIDRaw, _ := c.Get("user_id")
	currentUserID, _ := uuid.Parse(currentUserIDRaw.(string))

	// 2️⃣ Permission check - Only Admin, SUPERADMIN, and Manager can withdraw
	if role != "SUPERADMIN" && role != "ADMIN" && role != "MANAGER" {
		utils.RespondWithError(c, 403, "only SUPERADMIN, ADMIN, and MANAGER can withdraw approved leaves")
		return
	}

	// 2️⃣A Check if MANAGER has permission to withdraw leaves
	if role == "MANAGER" {
		hasPermission, err := h.Query.ChackManagerPermission()
		if err != nil {
			utils.RespondWithError(c, http.StatusInternalServerError, "failed to check manager permission")
			return
		}
		if !hasPermission {
			utils.RespondWithError(c, 403, "MANAGER does not have permission to withdraw leaves")
			return
		}
	}

	// 3️⃣ Parse Leave ID
	leaveIDStr := c.Param("id")
	leaveID, err := uuid.Parse(leaveIDStr)
	if err != nil {
		utils.RespondWithError(c, 400, "invalid leave ID")
		return
	}

	// 4️⃣ Parse optional reason from request body
	var input struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&input)

	// 5️⃣ Start transaction
	tx, err := h.Query.DB.Beginx()
	if err != nil {
		utils.RespondWithError(c, 500, "failed to start transaction")
		return
	}
	defer tx.Rollback()

	// 6️⃣ Fetch leave details
	var leave Leave
	err = tx.Get(&leave, `
		SELECT id, employee_id, leave_type_id, start_date, end_date, days, status, created_at
		FROM Tbl_Leave 
		WHERE id=$1 
		FOR UPDATE
	`, leaveID)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.RespondWithError(c, 404, "leave request not found")
			return
		}
		utils.RespondWithError(c, 500, "failed to fetch leave: "+err.Error())
		return
	}

	// 7️⃣ Prevent withdrawing own leave
	if leave.EmployeeID == currentUserID {
		utils.RespondWithError(c, 403, "you cannot withdraw your own leave. Please contact your manager or admin")
		return
	}

	// 8️⃣ MANAGER validation
	if role == "MANAGER" {
		// Verify reporting relationship
		var managerID uuid.UUID
		err := tx.Get(&managerID, "SELECT manager_id FROM Tbl_Employee WHERE id=$1", leave.EmployeeID)
		if err != nil {
			utils.RespondWithError(c, 500, "failed to verify manager relationship")
			return
		}
		if managerID != currentUserID {
			utils.RespondWithError(c, 403, "managers can only withdraw leaves of their team members")
			return
		}

		// Manager can only act on APPROVED leaves
		if leave.Status != "APPROVED" {
			utils.RespondWithError(c, 400, fmt.Sprintf("cannot withdraw leave with status: %s. Only approved leaves can be withdrawn", leave.Status))
			return
		}
	}

	// 9️⃣ ADMIN/SUPERADMIN validation
	if role == "ADMIN" || role == "SUPERADMIN" {
		// Admin can act on APPROVED or WITHDRAWAL_PENDING leaves
		if leave.Status != "APPROVED" && leave.Status != "WITHDRAWAL_PENDING" {
			utils.RespondWithError(c, 400, fmt.Sprintf("cannot withdraw leave with status: %s", leave.Status))
			return
		}
	}

	// ========================================
	// MANAGER WITHDRAWAL REQUEST (First Level)
	// ========================================
	if role == "MANAGER" {
		withdrawalReason := input.Reason
		if withdrawalReason == "" {
			withdrawalReason = "Withdrawal requested by Manager"
		}

		// Update status to WITHDRAWAL_PENDING
		_, err = tx.Exec(`
			UPDATE Tbl_Leave 
			SET status='WITHDRAWAL_PENDING', reason=$1, approved_by=$2, updated_at=NOW() 
			WHERE id=$3
		`, withdrawalReason, currentUserID, leaveID)
		if err != nil {
			utils.RespondWithError(c, 500, "failed to request withdrawal: "+err.Error())
			return
		}

		// Fetch employee and manager details
		var empDetails struct {
			Email    string `db:"email"`
			FullName string `db:"full_name"`
		}
		h.Query.DB.Get(&empDetails, "SELECT email, full_name FROM Tbl_Employee WHERE id=$1", leave.EmployeeID)

		var managerName string
		h.Query.DB.Get(&managerName, "SELECT full_name FROM Tbl_Employee WHERE id=$1", currentUserID)

		var leaveTypeName string
		h.Query.DB.Get(&leaveTypeName, "SELECT name FROM Tbl_Leave_type WHERE id=$1", leave.LeaveTypeID)

		tx.Commit()

		// Notify admins about pending withdrawal
		go func() {
			var adminEmails []string
			h.Query.DB.Select(&adminEmails, `
				SELECT e.email 
				FROM Tbl_Employee e
				JOIN Tbl_Role r ON e.role_id = r.id
				WHERE r.type IN ('ADMIN', 'SUPERADMIN') AND e.status = 'active'
			`)

			if len(adminEmails) > 0 {
				utils.SendLeaveWithdrawalPendingEmail(
					adminEmails,
					empDetails.FullName,
					leaveTypeName,
					leave.StartDate.Format("2006-01-02"),
					leave.EndDate.Format("2006-01-02"),
					leave.Days,
					managerName,
					withdrawalReason,
				)
			}
		}()

		c.JSON(200, gin.H{
			"message":           "withdrawal request submitted. Pending final approval from ADMIN/SUPERADMIN",
			"status":            "WITHDRAWAL_PENDING",
			"leave_id":          leaveID,
			"withdrawal_by":     currentUserID,
			"withdrawal_reason": withdrawalReason,
		})
		return
	}

	// ========================================
	// ADMIN/SUPERADMIN FINAL WITHDRAWAL (Second Level)
	// ========================================
	if role == "ADMIN" || role == "SUPERADMIN" {
		withdrawalReason := input.Reason
		if withdrawalReason == "" {
			withdrawalReason = fmt.Sprintf("Withdrawn by %s", role)
		}

		// Update status to WITHDRAWN
		_, err = tx.Exec(`
			UPDATE Tbl_Leave 
			SET status='WITHDRAWN', reason=$1, approved_by=$2, updated_at=NOW() 
			WHERE id=$3
		`, withdrawalReason, currentUserID, leaveID)
		if err != nil {
			utils.RespondWithError(c, 500, "failed to finalize withdrawal: "+err.Error())
			return
		}

		// Restore leave balance (reverse the deduction)
		_, err = tx.Exec(`
			UPDATE Tbl_Leave_balance 
			SET used = used - $1, closing = closing + $1, updated_at = NOW()
			WHERE employee_id=$2 AND leave_type_id=$3 AND year = EXTRACT(YEAR FROM CURRENT_DATE)
		`, leave.Days, leave.EmployeeID, leave.LeaveTypeID)
		if err != nil {
			utils.RespondWithError(c, 500, "failed to restore leave balance: "+err.Error())
			return
		}

		// Fetch data BEFORE committing transaction
		empDetails, err := h.Query.GetEmployeeDetailsForNotification(leave.EmployeeID)
		if err != nil {
			fmt.Printf("Failed to fetch employee details: %v\n", err)
		}

		var leaveTypeName string
		err = h.Query.DB.Get(&leaveTypeName, "SELECT name FROM Tbl_Leave_type WHERE id=$1", leave.LeaveTypeID)
		if err != nil {
			fmt.Printf("Failed to fetch leave type: %v\n", err)
		}

		var withdrawnByName string
		err = h.Query.DB.Get(&withdrawnByName, "SELECT full_name FROM Tbl_Employee WHERE id=$1", currentUserID)
		if err != nil {
			fmt.Printf("Failed to fetch withdrawn by name: %v\n", err)
		}
		admins, err := h.Query.GetAdminAndEmployeeEmail(leave.EmployeeID)
		if err != nil {
			fmt.Printf("Failed to fetch admin and employee emails: %v\n", err)
		}

		// Commit transaction
		err = tx.Commit()
		if err != nil {
			utils.RespondWithError(c, 500, "failed to commit transaction")
			return
		}

		// Send final withdrawal notification (async)
		if empDetails.Email != "" && leaveTypeName != "" && withdrawnByName != "" {
			go func(email, name, leaveType, startDate, endDate string, days float64, withdrawnBy, withdrawnRole, reason string) {
				fmt.Printf("📧 Sending withdrawal email to %s...\n", email)
				err := utils.SendLeaveWithdrawalEmail(
					admins,
					email,
					name,
					leaveType,
					startDate,
					endDate,
					days,
					withdrawnBy,
					withdrawnRole,
					reason,
				)
				if err != nil {
					fmt.Printf("❌ Failed to send withdrawal email: %v\n", err)
				} else {
					fmt.Printf("✅ Withdrawal email sent successfully to %s\n", email)
				}
			}(empDetails.Email, empDetails.FullName, leaveTypeName,
				leave.StartDate.Format("2006-01-02"),
				leave.EndDate.Format("2006-01-02"),
				leave.Days, withdrawnByName, role, input.Reason)
		}

		c.JSON(200, gin.H{
			"message":           "leave withdrawn successfully and balance restored",
			"status":            "WITHDRAWN",
			"leave_id":          leaveID,
			"days_restored":     leave.Days,
			"withdrawal_by":     currentUserID,
			"withdrawal_role":   role,
			"withdrawal_reason": withdrawalReason,
		})
		return
	}

	// Should not reach here
	utils.RespondWithError(c, 500, "unexpected error in leave withdrawal process")
}

// GetManagerLeaveHistory - GET /api/leaves/manager/history
// Manager gets leave history of their team members
func (h *HandlerFunc) GetManagerLeaveHistory(c *gin.Context) {
	// 1️⃣ Get current user info with validation
	role := c.GetString("role")
	if role == "" {
		utils.RespondWithError(c, http.StatusUnauthorized, "Role not found in context")
		return
	}

	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		utils.RespondWithError(c, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	currentUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid user ID format: "+err.Error())
		return
	}

	// 2️⃣ Permission check - Only MANAGER can use this endpoint
	if role != "MANAGER" {
		utils.RespondWithError(c, http.StatusForbidden, "Only managers can access team leave history")
		return
	}

	// 3️⃣ Query to get team members' leave history
	query := `
		SELECT 
			l.id,
			e.full_name AS employee,
			lt.name AS leave_type,
			l.start_date,
			l.end_date,
			l.days,
			COALESCE(l.reason, '') AS reason,
			l.status,
			l.created_at AS applied_at
		FROM Tbl_Leave l
		INNER JOIN Tbl_Employee e ON l.employee_id = e.id
		INNER JOIN Tbl_Leave_Type lt ON lt.id = l.leave_type_id
		WHERE e.manager_id = $1
		ORDER BY l.created_at DESC
	`

	// 4️⃣ Execute query with proper error handling
	var result []models.LeaveResponse
	err = h.Query.DB.Select(&result, query, currentUserID)
	if err != nil {
		// Log the error for debugging
		fmt.Printf("❌ GetManagerLeaveHistory DB Error: %v\n", err)
		fmt.Printf("Manager ID: %s\n", currentUserID)

		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch team leave history: "+err.Error())
		return
	}

	// 5️⃣ Handle empty result
	if result == nil {
		result = []models.LeaveResponse{}
	}

	// 6️⃣ Response with metadata
	c.JSON(http.StatusOK, gin.H{
		"message":      "Team leave history fetched successfully",
		"manager_id":   currentUserID,
		"total_leaves": len(result),
		"leaves":       result,
	})
}
