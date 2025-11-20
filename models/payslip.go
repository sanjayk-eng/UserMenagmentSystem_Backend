package models

import "github.com/go-oauth2/oauth2/utils/uuid"

//
// Payslip Model
// -----------------------------------------------------------------------------
// Stores salary computation details generated during payroll.
// Includes deductions, net salary, and generated PDF link.
//
type Payslip struct {
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`

	// Payroll Execution
	PayrollRunID uint       `json:"payroll_run_id" binding:"required"`
	PayrollRun   PayrollRun `gorm:"foreignKey:PayrollRunID"`

	// Employee Details
	EmployeeID uuid.UUID `gorm:"type:uuid;not null" json:"employee_id" binding:"required"`
	Employee   Employee  `gorm:"foreignKey:EmployeeID"`

	// Salary Breakdown
	BasicSalary     float64 `json:"basic_salary" binding:"required"`
	WorkingDays     int     `json:"working_days" binding:"required"`
	AbsentDays      int     `json:"absent_days"`
	DeductionAmount float64 `json:"deduction_amount"`
	NetSalary       float64 `json:"net_salary"`

	// Path to saved PDF payslip
	PDFPath string `gorm:"type:varchar(255)" json:"pdf_path"`
}
