package controllers

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/jung-kurt/gofpdf"
	"github.com/sanjayk-eng/UserMenagmentSystem_Backend/utils"
)

// PayrollPreview represents preview data for a payroll run
type PayrollPreview struct {
	EmployeeID  uuid.UUID `json:"employee_id"`
	Employee    string    `json:"employee"`
	BasicSalary float64   `json:"basic_salary"`
	WorkingDays int       `json:"working_days"`
	AbsentDays  int       `json:"absent_days"`
	Deductions  float64   `json:"deductions"`
	NetSalary   float64   `json:"net_salary"`
}

// RunPayroll handles payroll preview
func (h *HandlerFunc) RunPayroll(c *gin.Context) {
	roleRaw, _ := c.Get("role")
	role := roleRaw.(string)
	if role != "SUPERADMIN" && role != "ADMIN" && role != "HR" {
		utils.RespondWithError(c, 403, "Not authorized to run payroll")
		return
	}

	var input struct {
		Month int `json:"month" validate:"required"`
		Year  int `json:"year" validate:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondWithError(c, 400, "Invalid input: "+err.Error())
		return
	}

	now := time.Now()
	if input.Year > now.Year() || (input.Year == now.Year() && input.Month >=int(now.Month())) {
		utils.RespondWithError(c, http.StatusBadRequest, "Cannot run payroll for future months")
		return
	}


	// --- Check if payroll already exists and is finalized ---
	var existingStatus string
	err := h.Query.DB.Get(&existingStatus, `
        SELECT status 
        FROM Tbl_Payroll_run 
        WHERE month=$1 AND year=$2
    `, input.Month, input.Year)
	if err == nil && strings.ToUpper(strings.TrimSpace(existingStatus)) == "FINALIZED" {
		utils.RespondWithError(c, 400, "Payroll for this month and year is already finalized")
		return
	}

	// --- Fetch active employees ---
	var employees []struct {
		ID          uuid.UUID `db:"id"`
		FullName    string    `db:"full_name"`
		Salary      float64   `db:"salary"`
		Status      string    `db:"status"`
		JoiningDate time.Time `db:"joining_date"`
	}
	query := `
        SELECT id, full_name, salary, status, joining_date
        FROM Tbl_Employee
        WHERE status='active'
          AND (EXTRACT(YEAR FROM joining_date) < $1 
               OR (EXTRACT(YEAR FROM joining_date) = $1 AND EXTRACT(MONTH FROM joining_date) <= $2))
    `
	err = h.Query.DB.Select(&employees, query, input.Year, input.Month)
	if err != nil {
		utils.RespondWithError(c, 500, "Failed to fetch employees: "+err.Error())
		return
	}

	// --- Fetch working days ---
	var workingDays int
	err = h.Query.DB.Get(&workingDays, `
		SELECT working_days_per_month 
		FROM Tbl_Company_Settings 
		ORDER BY created_at DESC 
		LIMIT 1
	`)
	if err != nil {
		workingDays = 22 // fallback default
	}

	totalPayroll := 0.0
	totalDeductions := 0.0
	var previews []PayrollPreview

	for _, emp := range employees {
		// Calculate absent days for this specific month only
		// Handle cross-month leaves correctly
		absentDays := calculateAbsentDaysForMonth(h.Query.DB, emp.ID, input.Month, input.Year)
		if absentDays < 0 {
			utils.RespondWithError(c, 500, "Failed to calculate absent days")
			return
		}

		deduction := emp.Salary / float64(workingDays) * absentDays
		net := emp.Salary - deduction

		previews = append(previews, PayrollPreview{
			EmployeeID:  emp.ID,
			Employee:    emp.FullName,
			BasicSalary: emp.Salary,
			WorkingDays: workingDays,
			AbsentDays:  int(absentDays),
			Deductions:  deduction,
			NetSalary:   net,
		})

		totalPayroll += net
		totalDeductions += deduction
	}

	// --- Create payroll run record ---
	runID := uuid.New()
	_, err = h.Query.DB.Exec(
		`INSERT INTO Tbl_Payroll_run (id, month, year, status) VALUES ($1,$2,$3,$4)`,
		runID, input.Month, input.Year, "PREVIEW",
	)
	if err != nil {
		utils.RespondWithError(c, 500, "Failed to create payroll run: "+err.Error())
		return
	}

	c.JSON(200, gin.H{
		"payroll_run_id":   runID,
		"month":            input.Month,
		"year":             input.Year,
		"total_payroll":    totalPayroll,
		"total_deductions": totalDeductions,
		"employees_count":  len(employees),
		"payroll_preview":  previews,
	})
}

// FinalizePayroll - generates payslips
// Only SUPERADMIN can finalize payroll
func (h *HandlerFunc) FinalizePayroll(c *gin.Context) {
	// --- Role Check - Only SUPERADMIN ---
	roleRaw, _ := c.Get("role")
	role := roleRaw.(string)
	if role != "SUPERADMIN" {
		utils.RespondWithError(c, 403, "Only SUPERADMIN can finalize payroll")
		return
	}

	// --- Parse Payroll Run ID ---
	runID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondWithError(c, 400, "Invalid payroll run ID")
		return
	}

	// --- Fetch Payroll Run Data ---
	var run struct {
		Month  int    `db:"month"`
		Year   int    `db:"year"`
		Status string `db:"status"`
	}
	err = h.Query.DB.Get(&run,
		`SELECT month, year, status FROM Tbl_Payroll_run WHERE id=$1`, runID)
	if err != nil {
		utils.RespondWithError(c, 404, "Payroll run not found")
		return
	}

	// --- Block if Already Finalized ---
	if run.Status == "FINALIZED" {
		utils.RespondWithError(c, 400, "Payroll already finalized")
		return
	}

	// --- Fetch working days ---
	var workingDays int
	err = h.Query.DB.Get(&workingDays,
		`SELECT working_days_per_month FROM Tbl_Company_Settings ORDER BY created_at DESC LIMIT 1`)
	if err != nil || workingDays <= 0 {
		workingDays = 22 // fallback default
	}

	// --- Transaction Start ---
	tx, err := h.Query.DB.Beginx()
	if err != nil {
		utils.RespondWithError(c, 500, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	// --- Fetch Only Employees Belonging To The Payroll Run Period ---
	var employees []struct {
		ID       uuid.UUID `db:"id"`
		FullName string    `db:"full_name"`
		Salary   float64   `db:"salary"`
	}

	err = tx.Select(&employees, `
        SELECT e.id, e.full_name, e.salary
        FROM Tbl_Employee e
        JOIN Tbl_Payroll_run r ON r.id = $1
        WHERE e.status='active'
          AND (
               EXTRACT(YEAR FROM e.joining_date) < r.year
               OR (EXTRACT(YEAR FROM e.joining_date) = r.year
                   AND EXTRACT(MONTH FROM e.joining_date) <= r.month)
              )
	`, runID)

	if err != nil {
		utils.RespondWithError(c, 500, "Failed to fetch payroll employees: "+err.Error())
		return
	}

	// --- Generate Payslips ---
	var payslipIDs []uuid.UUID

	for _, emp := range employees {
		// Calculate absent days for this specific month only
		// Handle cross-month leaves correctly
		absentDays := calculateAbsentDaysForMonth(h.Query.DB, emp.ID, run.Month, run.Year)
		if absentDays < 0 {
			utils.RespondWithError(c, 500, "Failed to calculate absent days")
			return
		}

		deduction := emp.Salary / float64(workingDays) * absentDays
		net := emp.Salary - deduction

		pID := uuid.New()
		_, err = tx.Exec(`
			INSERT INTO Tbl_Payslip 
			(id, payroll_run_id, employee_id, basic_salary, working_days, absent_days, deduction_amount, net_salary)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		`, pID, runID, emp.ID, emp.Salary, workingDays, int(absentDays), deduction, net)

		if err != nil {
			utils.RespondWithError(c, 500, "Payslip insert failed: "+err.Error())
			return
		}

		payslipIDs = append(payslipIDs, pID)
	}

	// --- Mark Payroll Run Finalized ---
	_, err = tx.Exec(`UPDATE Tbl_Payroll_run SET status='FINALIZED', updated_at=NOW() WHERE id=$1`, runID)
	if err != nil {
		utils.RespondWithError(c, 500, "Failed to update payroll run: "+err.Error())
		return
	}

	if err = tx.Commit(); err != nil {
		utils.RespondWithError(c, 500, "Failed to commit: "+err.Error())
		return
	}

	// --- Success Response ---
	c.JSON(http.StatusOK, gin.H{
		"message":        "Payroll finalized successfully",
		"payroll_run_id": runID,
		"payslips":       payslipIDs,
	})
}

// GetPayslipPDF - GET /api/payroll/payslips/:id/pdf
func (h *HandlerFunc) GetPayslipPDF(c *gin.Context) {
	payslipID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid payslip ID"})
		return
	}

	var payslip struct {
		EmployeeID   uuid.UUID `db:"employee_id"`
		EmployeeName string    `db:"full_name"`
		Email        string    `db:"email"`
		Month        int       `db:"month"`
		Year         int       `db:"year"`
		BasicSalary  float64   `db:"basic_salary"`
		WorkingDays  int       `db:"working_days"`
		AbsentDays   int       `db:"absent_days"`
		Deductions   float64   `db:"deduction_amount"`
		NetSalary    float64   `db:"net_salary"`
	}

	err = h.Query.DB.Get(&payslip, `
		SELECT e.id as employee_id, e.full_name, e.email, 
		       p.basic_salary, p.working_days, p.absent_days, 
		       p.deduction_amount, p.net_salary,
		       pr.month, pr.year
		FROM Tbl_Payslip p
		JOIN Tbl_Employee e ON e.id = p.employee_id
		JOIN Tbl_Payroll_run pr ON pr.id = p.payroll_run_id
		WHERE p.id = $1
	`, payslipID)
	if err != nil {
		c.JSON(404, gin.H{"error": "Payslip not found: " + err.Error()})
		return
	}

	// Create PDF with improved design
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetAutoPageBreak(false, 0)

	// ========================================
	// HEADER SECTION - Company Branding
	// ========================================
	pdf.SetFillColor(41, 128, 185) // Professional blue
	pdf.Rect(0, 0, 210, 45, "F")

	pdf.SetTextColor(255, 255, 255) // White text
	pdf.SetFont("Arial", "B", 26)
	pdf.SetY(12)
	pdf.CellFormat(0, 10, "ZENITHIVE", "", 1, "C", false, 0, "")

	pdf.SetFont("Arial", "", 12)
	pdf.SetY(25)
	pdf.CellFormat(0, 6, "Payroll Management System", "", 1, "C", false, 0, "")

	pdf.SetFont("Arial", "B", 14)
	pdf.SetY(35)
	monthNames := []string{"", "January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December"}
	pdf.CellFormat(0, 6, fmt.Sprintf("Salary Slip - %s %d", monthNames[payslip.Month], payslip.Year), "", 1, "C", false, 0, "")

	// ========================================
	// EMPLOYEE INFORMATION SECTION
	// ========================================
	pdf.SetY(55)
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Arial", "B", 14)
	pdf.SetFillColor(236, 240, 241) // Light gray background
	pdf.CellFormat(0, 10, "  EMPLOYEE INFORMATION", "", 1, "L", true, 0, "")

	pdf.SetFont("Arial", "", 11)
	pdf.Ln(2)

	// Employee details in two columns
	leftX := 15.0
	rightX := 110.0
	currentY := pdf.GetY()

	// Left column
	pdf.SetXY(leftX, currentY)
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(40, 7, "Employee Name:")
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(0, 7, payslip.EmployeeName)

	currentY += 8
	pdf.SetXY(leftX, currentY)
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(40, 7, "Employee ID:")
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(0, 7, payslip.EmployeeID.String()[:8]+"...")

	// Right column
	currentY = pdf.GetY() - 8
	pdf.SetXY(rightX, currentY)
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(30, 7, "Email:")
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(0, 7, payslip.Email)

	currentY += 8
	pdf.SetXY(rightX, currentY)
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(30, 7, "Pay Period:")
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(0, 7, fmt.Sprintf("%s %d", monthNames[payslip.Month], payslip.Year))

	// ========================================
	// EARNINGS SECTION
	// ========================================
	pdf.SetY(currentY + 15)
	pdf.SetFont("Arial", "B", 14)
	pdf.SetFillColor(46, 204, 113) // Green
	pdf.SetTextColor(255, 255, 255)
	pdf.CellFormat(0, 10, "  EARNINGS", "", 1, "L", true, 0, "")

	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Arial", "B", 11)
	pdf.SetFillColor(232, 245, 233) // Light green
	pdf.CellFormat(130, 9, "  Description", "1", 0, "L", true, 0, "")
	pdf.CellFormat(50, 9, "Amount (INR)", "1", 1, "C", true, 0, "")

	pdf.SetFont("Arial", "", 11)
	pdf.CellFormat(130, 9, "  Basic Salary", "1", 0, "L", false, 0, "")
	pdf.CellFormat(50, 9, fmt.Sprintf("%.2f", payslip.BasicSalary), "1", 1, "R", false, 0, "")

	pdf.SetFont("Arial", "B", 11)
	pdf.SetFillColor(232, 245, 233)
	pdf.CellFormat(130, 9, "  GROSS EARNINGS", "1", 0, "L", true, 0, "")
	pdf.CellFormat(50, 9, fmt.Sprintf("%.2f", payslip.BasicSalary), "1", 1, "R", true, 0, "")

	// ========================================
	// DEDUCTIONS SECTION
	// ========================================
	pdf.Ln(5)
	pdf.SetFont("Arial", "B", 14)
	pdf.SetFillColor(231, 76, 60) // Red
	pdf.SetTextColor(255, 255, 255)
	pdf.CellFormat(0, 10, "  DEDUCTIONS", "", 1, "L", true, 0, "")

	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Arial", "B", 11)
	pdf.SetFillColor(255, 235, 238) // Light red
	pdf.CellFormat(130, 9, "  Description", "1", 0, "L", true, 0, "")
	pdf.CellFormat(50, 9, "Amount (INR)", "1", 1, "C", true, 0, "")

	pdf.SetFont("Arial", "", 11)
	pdf.CellFormat(130, 9, fmt.Sprintf("  Leave Deduction (%d absent days)", payslip.AbsentDays), "1", 0, "L", false, 0, "")
	pdf.CellFormat(50, 9, fmt.Sprintf("%.2f", payslip.Deductions), "1", 1, "R", false, 0, "")

	pdf.SetFont("Arial", "B", 11)
	pdf.SetFillColor(255, 235, 238)
	pdf.CellFormat(130, 9, "  TOTAL DEDUCTIONS", "1", 0, "L", true, 0, "")
	pdf.CellFormat(50, 9, fmt.Sprintf("%.2f", payslip.Deductions), "1", 1, "R", true, 0, "")

	// ========================================
	// NET SALARY SECTION (Highlighted)
	// ========================================
	pdf.Ln(5)
	pdf.SetFont("Arial", "B", 16)
	pdf.SetFillColor(52, 73, 94) // Dark blue
	pdf.SetTextColor(255, 255, 255)
	pdf.CellFormat(130, 12, "  NET SALARY (Take Home)", "1", 0, "L", true, 0, "")
	pdf.SetFont("Arial", "B", 18)
	pdf.CellFormat(50, 12, fmt.Sprintf("%.2f", payslip.NetSalary), "1", 1, "R", true, 0, "")

	// ========================================
	// ATTENDANCE SUMMARY
	// ========================================
	pdf.Ln(8)
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Arial", "B", 12)
	pdf.SetFillColor(241, 196, 15) // Yellow/Gold
	pdf.SetTextColor(0, 0, 0)
	pdf.CellFormat(0, 9, "  ATTENDANCE SUMMARY", "", 1, "L", true, 0, "")

	pdf.SetFont("Arial", "", 10)
	pdf.SetFillColor(255, 249, 230) // Light yellow
	pdf.CellFormat(60, 8, "  Total Working Days", "1", 0, "L", true, 0, "")
	pdf.CellFormat(60, 8, "  Days Present", "1", 0, "L", true, 0, "")
	pdf.CellFormat(60, 8, "  Days Absent", "1", 1, "L", true, 0, "")

	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(60, 8, fmt.Sprintf("  %d days", payslip.WorkingDays), "1", 0, "L", false, 0, "")
	pdf.CellFormat(60, 8, fmt.Sprintf("  %d days", payslip.WorkingDays-payslip.AbsentDays), "1", 0, "L", false, 0, "")
	pdf.CellFormat(60, 8, fmt.Sprintf("  %d days", payslip.AbsentDays), "1", 1, "L", false, 0, "")

	// ========================================
	// CALCULATION BREAKDOWN
	// ========================================
	pdf.Ln(8)
	pdf.SetFont("Arial", "B", 11)
	pdf.SetFillColor(189, 195, 199) // Gray
	pdf.CellFormat(0, 8, "  CALCULATION BREAKDOWN", "", 1, "L", true, 0, "")

	pdf.SetFont("Arial", "", 10)
	pdf.Ln(2)
	pdf.MultiCell(0, 6, fmt.Sprintf(
		"Per Day Salary = Basic Salary / Working Days = %.2f / %d = %.2f\n"+
			"Leave Deduction = Per Day Salary x Absent Days = %.2f x %d = %.2f\n"+
			"Net Salary = Basic Salary - Leave Deduction = %.2f - %.2f = %.2f",
		payslip.BasicSalary, payslip.WorkingDays, payslip.BasicSalary/float64(payslip.WorkingDays),
		payslip.BasicSalary/float64(payslip.WorkingDays), payslip.AbsentDays, payslip.Deductions,
		payslip.BasicSalary, payslip.Deductions, payslip.NetSalary,
	), "", "L", false)

	// ========================================
	// FOOTER
	// ========================================
	pdf.SetY(270)
	pdf.SetFont("Arial", "I", 9)
	pdf.SetTextColor(128, 128, 128)
	pdf.CellFormat(0, 5, "This is a computer-generated payslip and does not require a signature.", "", 1, "C", false, 0, "")
	pdf.CellFormat(0, 5, fmt.Sprintf("Generated on: %s", time.Now().Format("02-Jan-2006 15:04:05")), "", 1, "C", false, 0, "")

	pdf.SetDrawColor(41, 128, 185)
	pdf.SetLineWidth(0.5)
	pdf.Line(15, 285, 195, 285)

	// ========================================
	// SAVE PDF
	// ========================================
	os.MkdirAll("./tmp", os.ModePerm)
	pdfPath := fmt.Sprintf("./tmp/payslip_%s.pdf", payslipID)
	err = pdf.OutputFileAndClose(pdfPath)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to generate PDF: " + err.Error()})
		return
	}

	// Update PDF path in DB
	_, _ = h.Query.DB.Exec("UPDATE Tbl_Payslip SET pdf_path=$1, updated_at=NOW() WHERE id=$2", pdfPath, payslipID)

	// Serve PDF
	c.File(pdfPath)
}

func (h *HandlerFunc) GetFinalizedPayslips(c *gin.Context) {
	roleValue, ok := c.Get("role")
	if !ok {
		utils.RespondWithError(c, 500, "Failed to get role")
		return
	}
	role := roleValue.(string)

	var rows *sql.Rows
	var err error

	// 🌟 If Employee or Manager -> only their own slips
	if role == "EMPLOYEE" || role == "MANAGER" {
		empIDValue, ok := c.Get("user_id")
		if !ok {
			utils.RespondWithError(c, 500, "Failed to get employee ID")
			return
		}

		// empIDValue is string, parse it to uuid
		empIDStr, ok := empIDValue.(string)
		if !ok {
			utils.RespondWithError(c, 500, "Invalid employee ID format")
			return
		}

		empID, err := uuid.Parse(empIDStr)
		if err != nil {
			utils.RespondWithError(c, 500, "Failed to parse employee ID: "+err.Error())
			return
		}
		rows, err = h.Query.GetFinalizedPayslipsByEmployee(empID)
	} else {
		// 🌟 SuperAdmin / Admin -> all slips
		rows, err = h.Query.GetAllFinalizedPayslips()
	}

	if err != nil {
		utils.RespondWithError(c, 500, "Failed to fetch payslips: "+err.Error())
		return
	}
	
	// Only defer close if rows is not nil
	if rows != nil {
		defer rows.Close()
	} else {
		utils.RespondWithError(c, 500, "No rows returned from query")
		return
	}

	type FullPayslipResponse struct {
		PayslipID       uuid.UUID `json:"payslip_id"`
		EmployeeID      uuid.UUID `json:"employee_id"`
		FullName        string    `json:"full_name"`
		Email           string    `json:"email"`
		Month           int       `json:"month"`
		Year            int       `json:"year"`
		BasicSalary     float64   `json:"basic_salary"`
		WorkingDays     int       `json:"working_days"`
		AbsentDays      int       `json:"absent_days"`
		DeductionAmount float64   `json:"deduction_amount"`
		NetSalary       float64   `json:"net_salary"`
		PDFPath         string    `json:"pdf_path"`
		Calculation     string    `json:"calculation"`
		CreatedAt       string    `json:"created_at"`
	}

	var result []FullPayslipResponse

	for rows.Next() {
		var slip FullPayslipResponse
		err := rows.Scan(
			&slip.PayslipID,
			&slip.EmployeeID,
			&slip.FullName,
			&slip.Email,
			&slip.Month,
			&slip.Year,
			&slip.BasicSalary,
			&slip.WorkingDays,
			&slip.AbsentDays,
			&slip.DeductionAmount,
			&slip.NetSalary,
			&slip.PDFPath,
			&slip.Calculation,
			&slip.CreatedAt,
		)
		if err != nil {
			utils.RespondWithError(c, 500, "Scan failed: "+err.Error())
			return
		}
		result = append(result, slip)
	}

	if len(result) == 0 {
		c.JSON(200, gin.H{
			"message": "No finalized payslips found",
			"data":    []FullPayslipResponse{},
		})
		return
	}

	// 🎯 SUCCESS RESPONSE
	c.JSON(200, gin.H{
		"message":        "Finalized payslips fetched successfully",
		"total_payslips": len(result),
		"data":           result,
	})
}

// WithdrawPayslip - POST /api/payroll/payslips/:id/withdraw
// Two-level approval system for payslip withdrawal:
// 1. MANAGER approves → Status: MANAGER_APPROVED (no salary restoration)
// 2. ADMIN/SUPERADMIN finalizes → Status: WITHDRAWN (salary restored/adjusted)
func (h *HandlerFunc) WithdrawPayslip(c *gin.Context) {
	roleRaw, _ := c.Get("role")
	role := roleRaw.(string)

	if role == "EMPLOYEE" {
		utils.RespondWithError(c, 403, "Employees cannot withdraw payslips")
		return
	}

	approverIDRaw, _ := c.Get("user_id")
	approverID, _ := uuid.Parse(approverIDRaw.(string))

	payslipID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondWithError(c, 400, "Invalid payslip ID")
		return
	}

	var body struct {
		Action string `json:"action" validate:"required"` // APPROVE/REJECT
		Reason string `json:"reason"`                     // Optional reason for withdrawal
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

	tx, err := h.Query.DB.Beginx()
	if err != nil {
		utils.RespondWithError(c, 500, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	// Fetch payslip with status
	var payslip struct {
		ID              uuid.UUID  `db:"id"`
		EmployeeID      uuid.UUID  `db:"employee_id"`
		PayrollRunID    uuid.UUID  `db:"payroll_run_id"`
		BasicSalary     float64    `db:"basic_salary"`
		WorkingDays     int        `db:"working_days"`
		AbsentDays      int        `db:"absent_days"`
		DeductionAmount float64    `db:"deduction_amount"`
		NetSalary       float64    `db:"net_salary"`
		Status          *string    `db:"status"` // Can be NULL in old records
		WithdrawnBy     *uuid.UUID `db:"withdrawn_by"`
		WithdrawnReason *string    `db:"withdrawn_reason"`
	}

	err = tx.Get(&payslip, `
		SELECT id, employee_id, payroll_run_id, basic_salary, working_days, 
		       absent_days, deduction_amount, net_salary, status, withdrawn_by, withdrawn_reason
		FROM Tbl_Payslip 
		WHERE id=$1 
		FOR UPDATE
	`, payslipID)
	if err != nil {
		utils.RespondWithError(c, 404, "Payslip not found: "+err.Error())
		return
	}

	// 🔒 Prevent self-withdrawal
	if payslip.EmployeeID == approverID {
		utils.RespondWithError(c, 403, "You cannot withdraw your own payslip")
		return
	}

	// Default status to FINALIZED if NULL (for backward compatibility)
	currentStatus := "FINALIZED"
	if payslip.Status != nil {
		currentStatus = *payslip.Status
	}

	// 🔒 MANAGER validation
	if role == "MANAGER" {
		// Check manager permission setting
		exists, err := h.Query.ChackManagerPermission()
		if err != nil {
			utils.RespondWithError(c, 500, "Failed to get manager permission")
			return
		}
		if !exists {
			utils.RespondWithError(c, 403, "Manager withdrawal is not enabled")
			return
		}

		// Verify reporting relationship
		var managerID uuid.UUID
		err = tx.Get(&managerID, "SELECT manager_id FROM Tbl_Employee WHERE id=$1", payslip.EmployeeID)
		if err != nil {
			utils.RespondWithError(c, 500, "Failed to verify reporting relationship")
			return
		}

		if managerID != approverID {
			utils.RespondWithError(c, 403, "You can only withdraw payslips of employees who report to you")
			return
		}

		// Manager can only act on FINALIZED payslips
		if currentStatus != "FINALIZED" {
			utils.RespondWithError(c, 400, fmt.Sprintf("Cannot process payslip with status: %s", currentStatus))
			return
		}
	}

	// 🔒 ADMIN/SUPERADMIN validation
	if role == "ADMIN" || role == "SUPERADMIN" {
		// Admin can act on FINALIZED or MANAGER_APPROVED payslips
		if currentStatus != "FINALIZED" && currentStatus != "MANAGER_APPROVED" {
			utils.RespondWithError(c, 400, fmt.Sprintf("Cannot process payslip with status: %s", currentStatus))
			return
		}
	}

	// ========================================
	// REJECT ACTION
	// ========================================
	if body.Action == "REJECT" {
		// Reset status back to FINALIZED
		_, err = tx.Exec(`
			UPDATE Tbl_Payslip 
			SET status='FINALIZED', withdrawn_by=NULL, withdrawn_reason=NULL, updated_at=NOW() 
			WHERE id=$1
		`, payslipID)
		if err != nil {
			utils.RespondWithError(c, 500, "Failed to reject withdrawal: "+err.Error())
			return
		}

		tx.Commit()

		c.JSON(200, gin.H{
			"message": "Payslip withdrawal rejected successfully",
			"status":  "FINALIZED",
		})
		return
	}

	// ========================================
	// APPROVE ACTION
	// ========================================

	// MANAGER APPROVAL (First Level)
	if role == "MANAGER" {
		_, err = tx.Exec(`
			UPDATE Tbl_Payslip 
			SET status='MANAGER_APPROVED', withdrawn_by=$2, withdrawn_reason=$3, updated_at=NOW() 
			WHERE id=$1
		`, payslipID, approverID, body.Reason)
		if err != nil {
			utils.RespondWithError(c, 500, "Failed to approve withdrawal: "+err.Error())
			return
		}

		tx.Commit()

		c.JSON(200, gin.H{
			"message": "Payslip withdrawal approved by manager. Pending final approval from ADMIN/SUPERADMIN",
			"status":  "MANAGER_APPROVED",
		})
		return
	}

	// ADMIN/SUPERADMIN FINAL APPROVAL (Second Level)
	if role == "ADMIN" || role == "SUPERADMIN" {
		// Update status to WITHDRAWN
		_, err = tx.Exec(`
			UPDATE Tbl_Payslip 
			SET status='WITHDRAWN', withdrawn_by=$2, withdrawn_reason=$3, updated_at=NOW() 
			WHERE id=$1
		`, payslipID, approverID, body.Reason)
		if err != nil {
			utils.RespondWithError(c, 500, "Failed to finalize withdrawal: "+err.Error())
			return
		}

		// Fetch employee and payroll details for notification
		var empDetails struct {
			Email    string `db:"email"`
			FullName string `db:"full_name"`
		}
		h.Query.DB.Get(&empDetails, "SELECT email, full_name FROM Tbl_Employee WHERE id=$1", payslip.EmployeeID)

		var payrollDetails struct {
			Month int `db:"month"`
			Year  int `db:"year"`
		}
		h.Query.DB.Get(&payrollDetails, "SELECT month, year FROM Tbl_Payroll_run WHERE id=$1", payslip.PayrollRunID)

		tx.Commit()

		// Send withdrawal notification
		if empDetails.Email != "" {
			go func() {
				withdrawnByName := ""
				h.Query.DB.Get(&withdrawnByName, "SELECT full_name FROM Tbl_Employee WHERE id=$1", approverID)

				utils.SendPayslipWithdrawalEmail(
					empDetails.Email,
					empDetails.FullName,
					payrollDetails.Month,
					payrollDetails.Year,
					payslip.NetSalary,
					withdrawnByName,
					role,
					body.Reason,
				)
			}()
		}

		c.JSON(200, gin.H{
			"message": "Payslip withdrawn successfully",
			"status":  "WITHDRAWN",
		})
		return
	}

	// Should not reach here
	utils.RespondWithError(c, 500, "Unexpected error in payslip withdrawal process")
}


// calculateAbsentDaysForMonth calculates the number of absent days for a specific month
// Handles cross-month leaves correctly by counting only days within the payroll month
func calculateAbsentDaysForMonth(db *sqlx.DB, employeeID uuid.UUID, month, year int) float64 {
	// Get first and last day of the payroll month
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	lastDay := firstDay.AddDate(0, 1, -1) // Last day of the month

	// Fetch all approved UNPAID leaves that overlap with this month
	type LeaveRecord struct {
		StartDate time.Time `db:"start_date"`
		EndDate   time.Time `db:"end_date"`
		Days      float64   `db:"days"`
	}

	var leaves []LeaveRecord
	err := db.Select(&leaves, `
		SELECT l.start_date, l.end_date, l.days
		FROM Tbl_Leave l
		JOIN Tbl_Leave_type lt ON l.leave_type_id = lt.id
		WHERE l.employee_id=$1 
		AND l.status='APPROVED'
		AND lt.is_paid = false
		AND l.start_date <= $2
		AND l.end_date >= $3
	`, employeeID, lastDay, firstDay)

	if err != nil {
		fmt.Printf("Error fetching leaves: %v\n", err)
		return -1
	}

	// Calculate total absent days within this month
	totalAbsentDays := 0.0

	for _, leave := range leaves {
		// Determine the overlap period
		overlapStart := leave.StartDate
		if overlapStart.Before(firstDay) {
			overlapStart = firstDay
		}

		overlapEnd := leave.EndDate
		if overlapEnd.After(lastDay) {
			overlapEnd = lastDay
		}

		// Count working days in the overlap period
		daysInMonth := 0
		for d := overlapStart; !d.After(overlapEnd); d = d.AddDate(0, 0, 1) {
			// Skip weekends
			if d.Weekday() == time.Saturday || d.Weekday() == time.Sunday {
				continue
			}

			// Check if it's a holiday
			var isHoliday bool
			err := db.Get(&isHoliday, `
				SELECT EXISTS(SELECT 1 FROM Tbl_Holiday WHERE date=$1)
			`, d)
			if err == nil && !isHoliday {
				daysInMonth++
			}
		}

		totalAbsentDays += float64(daysInMonth)
	}

	return totalAbsentDays
}
