package models

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// ----------------- ROLE -----------------
type RoleInput struct {
	Type string `json:"type" validate:"required"`
}

// ----------------- EMPLOYEE -----------------
type EmployeeInput struct {
	ID          *uuid.UUID `json:"id,omitempty"` // optional UUID
	FullName    string     `json:"full_name" validate:"required"`
	Email       string     `json:"email" validate:"required,email"`
	Role        string     `json:"role" validate:"required"`
	Password    string     `json:"password" validate:"required"`
	ManagerID   *uuid.UUID `json:"manager_id,omitempty"`   // optional UUID
	Salary      *float64   `json:"salary,omitempty"`       // optional
	JoiningDate *time.Time `json:"joining_date,omitempty"` // optional
	Status      *string    `json:"status,omitempty"`       // optional, new field
	CreatedAt   *time.Time `json:"created_at,omitempty"`   // optional
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`   // optional
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
	ManagerName *string    `json:"manager_name,omitempty"` // optional
}

// ----------------- LEAVE TYPE -----------------
type LeaveType struct {
	ID                 int    `json:"id" db:"id"`
	Name               string `json:"name" db:"name"`
	IsPaid             bool   `json:"is_paid" db:"is_paid"`
	DefaultEntitlement int    `json:"default_entitlement" db:"default_entitlement"`
	// LeaveCount         int       `json:"leave_count" db:"leave_count"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type LeaveTypeInput struct {
	Name               string `json:"name" validate:"required"`
	IsPaid             *bool  `json:"is_paid,omitempty"`
	DefaultEntitlement *int   `json:"default_entitlement,omitempty"`
	LeaveCount         *int   `json:"leave_count,omitempty" validate:"omitempty,gt=0"`
}

// ----------------- LEAVE -----------------
type LeaveInput struct {
	EmployeeID   uuid.UUID  `json:"employee_id" validate:"required"`
	LeaveTypeID  int        `json:"leave_type_id" validate:"required"`
	StartDate    time.Time  `json:"start_date" validate:"required"`
	EndDate      time.Time  `json:"end_date" validate:"required"`
	Days         *float64   `json:"days,omitempty"`
	Status       string     `json:"status,omitempty"`
	AppliedByID  *uuid.UUID `json:"applied_by,omitempty"`
	ApprovedByID *uuid.UUID `json:"approved_by,omitempty"`
}

// ----------------- LEAVE BALANCE -----------------
type LeaveBalanceInput struct {
	EmployeeID  uuid.UUID `json:"employee_id" validate:"required"`
	LeaveTypeID int       `json:"leave_type_id" validate:"required"`
	Year        int       `json:"year,omitempty"`
	Opening     *float64  `json:"opening,omitempty"`
	Accrued     *float64  `json:"accrued,omitempty"`
	Used        *float64  `json:"used,omitempty"`
	Adjusted    *float64  `json:"adjusted,omitempty"`
	Closing     *float64  `json:"closing,omitempty"`
}

type LeaveResponse struct {
	ID           uuid.UUID `json:"id" db:"id"`                    // leave ID
	Employee     string    `json:"employee" db:"employee"`        // employee full name
	LeaveType    string    `json:"leave_type" db:"leave_type"`    // leave type name
	StartDate    time.Time `json:"start_date" db:"start_date"`    // leave start date
	EndDate      time.Time `json:"end_date" db:"end_date"`        // leave end date
	Days         int       `json:"days" db:"days"`                // leave days
	Status       string    `json:"status" db:"status"`            // leave status (Pending/Approved/Rejected)
	ApplyingDate time.Time `json:"applying_date" db:"applied_at"` // when leave was applied
}

// ----------------- LEAVE ADJUSTMENT -----------------
type LeaveAdjustmentInput struct {
	EmployeeID  uuid.UUID `json:"employee_id" validate:"required"`
	LeaveTypeID int       `json:"leave_type_id" validate:"required"`
	Quantity    float64   `json:"quantity" validate:"required"`
	Reason      *string   `json:"reason,omitempty"`
	CreatedByID uuid.UUID `json:"created_by" validate:"required"`
}

// ----------------- PAYROLL RUN -----------------
type PayrollRunInput struct {
	Month  int     `json:"month" validate:"required"`
	Year   int     `json:"year" validate:"required"`
	Status *string `json:"status,omitempty"`
}

// ----------------- PAYSLIP -----------------
type PayslipInput struct {
	PayrollRunID    uuid.UUID `json:"payroll_run_id" validate:"required"`
	EmployeeID      uuid.UUID `json:"employee_id" validate:"required"`
	BasicSalary     *float64  `json:"basic_salary,omitempty"`
	WorkingDays     *int      `json:"working_days,omitempty"`
	AbsentDays      *int      `json:"absent_days,omitempty"`
	DeductionAmount *float64  `json:"deduction_amount,omitempty"`
	NetSalary       *float64  `json:"net_salary,omitempty"`
	PdfPath         *string   `json:"pdf_path,omitempty"`
}
type PayrollEmployeeResponse struct {
	EmployeeName string  `json:"employee_name"`
	BasicSalary  float64 `json:"basic_salary"`
	WorkingDays  float64 `json:"working_days"` // float64 expected
	AbsentDays   float64 `json:"absent_days"`
	Deductions   float64 `json:"deductions"`
	NetSalary    float64 `json:"net_salary"`
}

// -------------------Loing input-----------------------
type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// ----------------- AUDIT -----------------
type AuditInput struct {
	ActorID  uuid.UUID  `json:"actor_id" validate:"required"`
	Action   *string    `json:"action,omitempty"`
	Entity   *string    `json:"entity,omitempty"`
	EntityID *uuid.UUID `json:"entity_id,omitempty"`
	Metadata *string    `json:"metadata,omitempty"` // JSON as string
}

type Holiday struct {
	ID        int64     `json:"id" db:"id"`
	Name      string    `json:"name" db:"name" binding:"required"`
	Date      time.Time `json:"date" db:"date" binding:"required"` // Input by user
	Day       string    `json:"day" db:"day"`                      // Automatically calculated
	Type      string    `json:"type" db:"type"`                    // Default "HOLIDAY"
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type FullPayslipResponse struct {
	PayslipID       uuid.UUID `json:"payslip_id"`
	EmployeeID      uuid.UUID `json:"employee_id"`
	FullName        string    `json:"full_name"`
	Email           string    `json:"email"`
	Month           int       `json:"month"` // from Payroll_Run
	Year            int       `json:"year"`
	BasicSalary     float64   `json:"basic_salary"`
	WorkingDays     int       `json:"working_days"`
	AbsentDays      int       `json:"absent_days"`
	DeductionAmount float64   `json:"deduction_amount"`
	NetSalary       float64   `json:"net_salary"`
	PDFPath         string    `json:"pdf_path"`
	CalculationText string    `json:"calculation"`
	CreatedAt       string    `json:"created_at"`
}

var Validate *validator.Validate

func InitValidator() {
	Validate = validator.New()
}
