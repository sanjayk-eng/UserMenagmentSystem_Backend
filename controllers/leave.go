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
	"github.com/sanjayk-eng/UserMenagmentSystem_Backend/utils"
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
func (h *HandlerFunc) ApplyLeave(c *gin.Context) {
	// Extract Employee Info from Middleware
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

	//role, ok := c.Get("role")
	// if !ok  {
	// 	utils.RespondWithError(c, http.StatusForbidden, "Only employees can apply leave")
	// 	return
	// }

	// Validate Employee Status
	var empStatus string
	err = h.Query.DB.Get(&empStatus, `
		SELECT status FROM Tbl_Employee WHERE id=$1
	`, employeeID)

	if err != nil {
		utils.RespondWithError(c, 500, "Failed to verify employee status")
		return
	}

	if empStatus == "deactive" {
		utils.RespondWithError(c, 403, "Your account is deactivated. You cannot apply leave")
		return
	}

	// Bind Input JSON
	var input models.LeaveInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid input: "+err.Error())
		return
	}

	input.EmployeeID = employeeID

	// Validate reason - Enhanced validation
	if input.Reason == "" {
		utils.RespondWithError(c, 400, "Leave reason is required. Please provide a valid reason for your leave request")
		return
	}
	
	// Trim whitespace and check minimum length
	input.Reason = strings.TrimSpace(input.Reason)
	if len(input.Reason) < 10 {
		utils.RespondWithError(c, 400, "Leave reason must be at least 10 characters long. Please provide a detailed reason")
		return
	}
	
	if len(input.Reason) > 500 {
		utils.RespondWithError(c, 400, "Leave reason is too long. Maximum 500 characters allowed")
		return
	}

	// Validate Date
	today := time.Now().Truncate(24 * time.Hour)
	if input.StartDate.Before(today) {
		utils.RespondWithError(c, 400, "Start date cannot be earlier than today")
		return
	}

	if input.EndDate.Before(input.StartDate) {
		utils.RespondWithError(c, 400, "End date cannot be earlier than start date")
		return
	}

	// Start Transaction
	tx, err := h.Query.DB.Beginx()
	if err != nil {
		utils.RespondWithError(c, 500, "Failed to start transaction")
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Calculate Working Days
	leaveDays, err := CalculateWorkingDays(tx, input.StartDate, input.EndDate)
	if err != nil {
		utils.RespondWithError(c, 500, "Failed to calculate leave days: "+err.Error())
		return
	}
	if leaveDays <= 0 {
		utils.RespondWithError(c, 400, "Leave days must be greater than 0")
		return
	}
	input.Days = &leaveDays

	// Validate Leave Type
	var leaveType struct {
		DefaultEntitlement int `db:"default_entitlement"`
	}
	err = tx.Get(&leaveType,
		"SELECT default_entitlement FROM Tbl_Leave_type WHERE id=$1",
		input.LeaveTypeID,
	)
	if err == sql.ErrNoRows {
		utils.RespondWithError(c, 400, "Invalid leave type")
		return
	}
	if err != nil {
		utils.RespondWithError(c, 500, "Failed to fetch leave type")
		return
	}

	// Get/Create Leave Balance
	var balance float64
	err = tx.Get(&balance, `
		SELECT closing 
		FROM Tbl_Leave_balance 
		WHERE employee_id=$1 AND leave_type_id=$2 
		AND year = EXTRACT(YEAR FROM CURRENT_DATE)
	`, employeeID, input.LeaveTypeID)

	if err == sql.ErrNoRows {
		balance = float64(leaveType.DefaultEntitlement)

		_, err = tx.Exec(`
			INSERT INTO Tbl_Leave_balance 
				(employee_id, leave_type_id, year, opening, accrued, used, adjusted, closing)
			VALUES ($1, $2, EXTRACT(YEAR FROM CURRENT_DATE), $3, 0, 0, 0, $3)
		`, employeeID, input.LeaveTypeID, leaveType.DefaultEntitlement)
		if err != nil {
			utils.RespondWithError(c, 500, "Failed to create leave balance")
			return
		}
	} else if err != nil {
		utils.RespondWithError(c, 500, "Failed to fetch leave balance")
		return
	}

	// Check Enough Leaves
	if balance < leaveDays {
		utils.RespondWithError(c, 400, "Insufficient leave balance")
		return
	}

	// Overlapping Leave Check - Only check Pending and Approved leaves
	var overlappingLeaves []struct {
		ID        uuid.UUID `db:"id"`
		LeaveType string    `db:"leave_type"`
		StartDate time.Time `db:"start_date"`
		EndDate   time.Time `db:"end_date"`
		Status    string    `db:"status"`
	}
	
	err = tx.Select(&overlappingLeaves, `
		SELECT l.id, lt.name as leave_type, l.start_date, l.end_date, l.status
		FROM Tbl_Leave l
		JOIN Tbl_Leave_type lt ON l.leave_type_id = lt.id
		WHERE l.employee_id=$1 
		AND l.status IN ('Pending','APPROVED')
		AND l.start_date <= $2 
		AND l.end_date >= $3
	`, employeeID, input.EndDate, input.StartDate)

	if err != nil {
		utils.RespondWithError(c, 500, "Failed to check overlapping leave")
		return
	}
	
	if len(overlappingLeaves) > 0 {
		// Build detailed error message
		overlap := overlappingLeaves[0]
		utils.RespondWithError(c, 400, 
			fmt.Sprintf("Overlapping leave exists: %s from %s to %s (Status: %s). Please cancel or modify the existing leave first",
				overlap.LeaveType,
				overlap.StartDate.Format("2006-01-02"),
				overlap.EndDate.Format("2006-01-02"),
				overlap.Status))
		return
	}

	// INSERT Leave with Reason
	var leaveID uuid.UUID
	err = tx.QueryRow(`
		INSERT INTO Tbl_Leave 
		(employee_id, leave_type_id, start_date, end_date, days, status, reason)
		VALUES ($1,$2,$3,$4,$5,'Pending',$6)
		RETURNING id
	`,
		employeeID,
		input.LeaveTypeID,
		input.StartDate,
		input.EndDate,
		leaveDays,
		input.Reason,
	).Scan(&leaveID)

	if err != nil {
		utils.RespondWithError(c, 500, "Failed to apply leave: "+err.Error())
		return
	}

	// Commit Transaction
	err = tx.Commit()
	if err != nil {
		utils.RespondWithError(c, 500, "Failed to commit transaction")
		return
	}

	// Response
	c.JSON(200, gin.H{
		"message":  "Leave applied successfully",
		"leave_id": leaveID,
		"days":     leaveDays,
		"reason":   input.Reason,
	})
}

// AdminAddLeave - POST /api/leaves/admin-add
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
	var leaves []models.LeaveType

	query := `SELECT id, name, is_paid, default_entitlement,  created_at, updated_at FROM Tbl_Leave_type ORDER BY id`
	err := s.Query.DB.Select(&leaves, query)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch leave types: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, leaves) // send models directly
}

// ActionLeave - POST /api/leaves/:id/action
func (s *HandlerFunc) ActionLeave(c *gin.Context) {
	fmt.Println("=============== run")
	roleRaw, _ := c.Get("role")
	role := roleRaw.(string)

	if role == "EMPLOYEE" {
		utils.RespondWithError(c, 403, "Employees cannot approve leaves")
		return
	}

	approverIDRaw, _ := c.Get("user_id")
	approverID, _ := uuid.Parse(approverIDRaw.(string))

	leaveID, err := uuid.Parse(c.Param("id"))
	fmt.Println(leaveID)
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
	defer tx.Rollback() // rollback if any error

	var leave Leave
	err = tx.Get(&leave, `SELECT * FROM Tbl_Leave WHERE id=$1 FOR UPDATE`, leaveID)
	if err != nil {
		utils.RespondWithError(c, 404, "Leave not found"+err.Error())
		return
	}

	if leave.Status != "Pending" {
		utils.RespondWithError(c, 400, "Leave already processed")
		return
	}

	if body.Action == "REJECT" {
		_, err = tx.Exec(`UPDATE Tbl_Leave SET status='REJECTED', approved_by=$2, updated_at=NOW() WHERE id=$1`, leaveID, approverID)
		if err != nil {
			utils.RespondWithError(c, 500, "Failed to reject leave: "+err.Error())
			return
		}
		tx.Commit()

		// Send rejection notification to employee
		go func() {
			var empDetails struct {
				Email    string `db:"email"`
				FullName string `db:"full_name"`
			}
			s.Query.DB.Get(&empDetails, "SELECT email, full_name FROM Tbl_Employee WHERE id=$1", leave.EmployeeID)

			var leaveTypeName string
			s.Query.DB.Get(&leaveTypeName, "SELECT name FROM Tbl_Leave_type WHERE id=$1", leave.LeaveTypeID)

			err := utils.SendLeaveRejectionEmail(
				empDetails.Email,
				empDetails.FullName,
				leaveTypeName,
				leave.StartDate.Format("2006-01-02"),
				leave.EndDate.Format("2006-01-02"),
				leave.Days,
			)
			if err != nil {
				fmt.Printf("Failed to send rejection email: %v\n", err)
			}
		}()

		c.JSON(200, gin.H{"message": "Leave rejected"})
		return
	}

	// APPROVE - First check balance
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

	// Check if sufficient balance exists
	if currentBalance < leave.Days {
		utils.RespondWithError(c, 400, fmt.Sprintf("Cannot approve: Insufficient leave balance. Available: %.1f days, Required: %.1f days", currentBalance, leave.Days))
		return
	}

	_, err = tx.Exec(`UPDATE Tbl_Leave SET status='APPROVED', approved_by=$2, updated_at=NOW() WHERE id=$1`, leaveID, approverID)
	if err != nil {
		utils.RespondWithError(c, 500, "Failed to approve leave: "+err.Error())
		return
	}

	_, err = tx.Exec(`UPDATE Tbl_Leave_balance SET used = used + $3, closing = closing - $3, updated_at = NOW() WHERE employee_id=$1 AND leave_type_id=$2`,
		leave.EmployeeID, leave.LeaveTypeID, leave.Days)
	if err != nil {
		utils.RespondWithError(c, 500, "Failed to update leave balance: "+err.Error())
		return
	}

	tx.Commit()

	// Send approval notification to employee
	go func() {
		var empDetails struct {
			Email    string `db:"email"`
			FullName string `db:"full_name"`
		}
		s.Query.DB.Get(&empDetails, "SELECT email, full_name FROM Tbl_Employee WHERE id=$1", leave.EmployeeID)

		var leaveTypeName string
		s.Query.DB.Get(&leaveTypeName, "SELECT name FROM Tbl_Leave_type WHERE id=$1", leave.LeaveTypeID)

		err := utils.SendLeaveApprovalEmail(
			empDetails.Email,
			empDetails.FullName,
			leaveTypeName,
			leave.StartDate.Format("2006-01-02"),
			leave.EndDate.Format("2006-01-02"),
			leave.Days,
		)
		if err != nil {
			fmt.Printf("Failed to send approval email: %v\n", err)
		}
	}()

	c.JSON(200, gin.H{"message": "Leave approved successfully"})
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
	// 1️⃣ Get Role & User ID
	role := c.GetString("role")
	userID, _ := uuid.Parse(c.GetString("user_id"))

	// 2️⃣ Base Query
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
	JOIN Tbl_Employee e ON l.employee_id = e.id
	JOIN Tbl_Leave_Type lt ON lt.id = l.leave_type_id
	`

	var (
		conditions []string
		args       []any
	)

	// 3️⃣ Role-based conditions
	if role == "EMPLOYEE" {
		conditions = append(conditions, "l.employee_id = $1")
		args = append(args, userID)
	}

	if role == "MANAGER" {
		conditions = append(conditions, "e.manager_id = $1")
		args = append(args, userID)
	}

	// Apply conditions safely
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY l.created_at DESC"

	// 4️⃣ Execute
	var result []models.LeaveResponse
	err := h.Query.DB.Select(&result, query, args...)
	if err != nil {
		utils.RespondWithError(c, 500, "Failed to fetch leaves: "+err.Error())
		return
	}

	// 5️⃣ Return success
	c.JSON(200, gin.H{
		"total": len(result),
		"data":  result,
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
		"message": "Leave cancelled successfully",
		"leave_id": leaveID,
	})
}

// WithdrawApprovedLeave - POST /api/leaves/:id/withdraw
// Allows admins/managers to withdraw an approved leave and restore balance
func (h *HandlerFunc) WithdrawApprovedLeave(c *gin.Context) {
	// Get user info from middleware
	roleRaw, _ := c.Get("role")
	role := roleRaw.(string)

	// Only admins and managers can withdraw approved leaves
	if role == "EMPLOYEE" {
		utils.RespondWithError(c, 403, "Only admins and managers can withdraw approved leaves")
		return
	}

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

	// Check if leave is approved
	if leave.Status != "APPROVED" {
		utils.RespondWithError(c, 400, fmt.Sprintf("Cannot withdraw leave with status: %s. Only approved leaves can be withdrawn", leave.Status))
		return
	}

	// Check if leave has already started
	today := time.Now().Truncate(24 * time.Hour)
	if leave.StartDate.Before(today) {
		utils.RespondWithError(c, 400, "Cannot withdraw leave that has already started or passed")
		return
	}

	// Update leave status to CANCELLED
	_, err = tx.Exec(`
		UPDATE Tbl_Leave 
		SET status='CANCELLED', updated_at=NOW() 
		WHERE id=$1
	`, leaveID)
	if err != nil {
		utils.RespondWithError(c, 500, "Failed to withdraw leave: "+err.Error())
		return
	}

	// Restore leave balance
	_, err = tx.Exec(`
		UPDATE Tbl_Leave_balance 
		SET used = used - $1, closing = closing + $1, updated_at = NOW()
		WHERE employee_id=$2 AND leave_type_id=$3
	`, leave.Days, leave.EmployeeID, leave.LeaveTypeID)
	if err != nil {
		utils.RespondWithError(c, 500, "Failed to restore leave balance: "+err.Error())
		return
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		utils.RespondWithError(c, 500, "Failed to commit transaction")
		return
	}

	// Send notification to employee
	go func() {
		var empDetails struct {
			Email    string `db:"email"`
			FullName string `db:"full_name"`
		}
		h.Query.DB.Get(&empDetails, "SELECT email, full_name FROM Tbl_Employee WHERE id=$1", leave.EmployeeID)

		var leaveTypeName string
		h.Query.DB.Get(&leaveTypeName, "SELECT name FROM Tbl_Leave_type WHERE id=$1", leave.LeaveTypeID)

		// You can create a new email template for this
		body := fmt.Sprintf(`
			<h2>Leave Withdrawn</h2>
			<p>Dear %s,</p>
			<p>Your approved leave has been withdrawn by management:</p>
			<ul>
				<li><strong>Leave Type:</strong> %s</li>
				<li><strong>Dates:</strong> %s to %s</li>
				<li><strong>Days:</strong> %.1f</li>
			</ul>
			<p>Your leave balance has been restored.</p>
			<p>Please contact your manager for more details.</p>
		`, empDetails.FullName, leaveTypeName, 
			leave.StartDate.Format("2006-01-02"), 
			leave.EndDate.Format("2006-01-02"), 
			leave.Days)

		utils.SendEmail(empDetails.Email, "Leave Withdrawn", body)
	}()

	c.JSON(200, gin.H{
		"message": "Leave withdrawn successfully and balance restored",
		"leave_id": leaveID,
		"days_restored": leave.Days,
	})
}
