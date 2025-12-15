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
	ID              *uuid.UUID `json:"id,omitempty"` // optional UUID
	FullName        string     `json:"full_name" validate:"required"`
	Email           string     `json:"email" validate:"required,email"`
	Role            string     `json:"role" validate:"required"`
	Password        string     `json:"password,omitempty"`       // optional - auto-generated if not provided
	ManagerID       *uuid.UUID `json:"manager_id,omitempty"`     // optional UUID
	DesignationID   *uuid.UUID `json:"designation_id,omitempty"` // optional UUID
	Salary          *float64   `json:"salary,omitempty"`         // optional
	JoiningDate     *time.Time `json:"joining_date,omitempty"`   // optional
	EndingDate      *time.Time `json:"ending_date,omitempty"`    // optional
	Status          *string    `json:"status,omitempty"`         // optional, new field
	CreatedAt       *time.Time `json:"created_at,omitempty"`     // optional
	UpdatedAt       *time.Time `json:"updated_at,omitempty"`     // optional
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
	ManagerName     *string    `json:"manager_name,omitempty"`     // optional
	DesignationName *string    `json:"designation_name,omitempty"` // optional
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
	EmployeeID    uuid.UUID  `json:"employee_id" validate:"required"`
	LeaveTypeID   int        `json:"leave_type_id" validate:"required"`
	LeaveTimingID *int       `json:"leave_timing_id,omitempty"` // Optional timing ID (defaults to 3 - Full Day)
	StartDate     time.Time  `json:"start_date" validate:"required"`
	EndDate       time.Time  `json:"end_date" validate:"required"`
	Reason        string     `json:"reason" validate:"required,min=10,max=500"` // Enhanced validation
	Days          *float64   `json:"days,omitempty"`
	Status        string     `json:"status,omitempty"`
	AppliedByID   *uuid.UUID `json:"applied_by,omitempty"`
	ApprovedByID  *uuid.UUID `json:"approved_by,omitempty"`
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
	AbsentDays      *float64  `json:"absent_days,omitempty"`
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
	AbsentDays      float64   `json:"absent_days"`
	DeductionAmount float64   `json:"deduction_amount"`
	NetSalary       float64   `json:"net_salary"`
	PDFPath         string    `json:"pdf_path"`
	CalculationText string    `json:"calculation"`
	CreatedAt       string    `json:"created_at"`
}
type LeaveResponse struct {
	ID              uuid.UUID `db:"id" json:"id"`
	Employee        string    `db:"employee" json:"employee"`
	LeaveType       string    `db:"leave_type" json:"leave_type"`
	IsPaid          bool      `db:"is_paid" json:"is_paid"`
	LeaveTimingType string    `db:"leave_timing_type" json:"leave_timing_type"`
	LeaveTiming     string    `db:"leave_timing" json:"leave_timing"`
	StartDate       time.Time `db:"start_date" json:"start_date"`
	EndDate         time.Time `db:"end_date" json:"end_date"`
	Days            float64   `db:"days" json:"days"`
	Reason          string    `db:"reason" json:"reason"`
	Status          string    `db:"status" json:"status"`
	AppliedAt       time.Time `db:"applied_at" json:"applied_at"`
}

var Validate *validator.Validate

func InitValidator() {
	Validate = validator.New()
}

// ----------------- DESIGNATION -----------------
type Designation struct {
	ID              string  `json:"id" db:"id"`
	DesignationName string  `json:"designation_name" db:"designation_name"`
	Description     *string `json:"description,omitempty" db:"description"`
}

type DesignationInput struct {
	DesignationName string  `json:"designation_name" validate:"required"`
	Description     *string `json:"description,omitempty"`
}

// CompanySettings struct mapping the DB table
type CompanySettings struct {
	ID                   uuid.UUID `db:"id" json:"id"`
	WorkingDaysPerMonth  int       `db:"working_days_per_month" json:"working_days_per_month"`
	AllowManagerAddLeave bool      `db:"allow_manager_add_leave" json:"allow_manager_add_leave"`
	CreatedAt            string    `db:"created_at" json:"created_at"`
	UpdatedAt            string    `db:"updated_at" json:"updated_at"`
}

type CompanyField struct {
	WorkingDaysPerMonth  int  `json:"working_days_per_month" binding:"required"`
	AllowManagerAddLeave bool `json:"allow_manager_add_leave"`
}

// ----------------- LOG -----------------
type LogResponse struct {
	ID        int       `json:"id" db:"id"`
	UserName  string    `json:"user_name" db:"user_name"`
	Action    string    `json:"action" db:"action"`
	Component string    `json:"component" db:"component"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Leave struct {
	ID            uuid.UUID  `db:"id"`
	EmployeeID    uuid.UUID  `db:"employee_id"`
	LeaveTypeID   int        `db:"leave_type_id"`
	LeaveTimingID *int       `db:"half_id"` // Timing ID (references Tbl_Half)
	StartDate     time.Time  `db:"start_date"`
	EndDate       time.Time  `db:"end_date"`
	Days          float64    `db:"days"`
	Status        string     `db:"status"`
	AppliedByID   *uuid.UUID `db:"applied_by"`
	ApprovedByID  *uuid.UUID `db:"approved_by"`
	Reason        string     `db:"reason"`
	CreatedAt     time.Time  `db:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at"`
}

// Leave Timing
type LeaveTimingResponse struct {
	ID        int        `json:"id" db:"id"`
	Type      string     `json:"type" db:"type"`
	Timing    string     `json:"timing" db:"timing"`
	CreatedAt *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt *time.Time `json:"updated_at" db:"updated_at"`
}
type UpdateLeaveTimingReq struct {
	ID     int    `uri:"id" validate:"required,oneof=1 2 3"`
	Timing string `json:"timing" validate:"required"`
}

type GetLeaveTimingByIDReq struct {
	ID int `uri:"id" validate:"required,oneof=1 2 3"`
}

// EQUIPMENT

type EquipmentCategoryRequest struct {
	Name        string  `json:"name" validate:"required,min=2,max=50"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=255"`
}
type EquipmentCategoryRes struct {
	ID          *string   `db:"id" json:"id,omitempty" validate:"omitempty,uuid4"`
	Name        string    `db:"name" json:"name" validate:"required,min=2,max=50"`
	Description string    `db:"description" json:"description,omitempty" validate:"omitempty,max=255"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

type EquipmentRequest struct {
	ID            *string `json:"id,omitempty" validate:"omitempty,uuid4"`
	Name          string  `json:"name" validate:"required,min=2,max=100"`
	CategoryID    string  `json:"category_id" validate:"required,uuid4"`
	Ownership     string  `json:"ownership,omitempty" validate:"omitempty,oneof=COMPANY SELF"`
	IsShared      *bool   `json:"is_shared,omitempty"`
	TotalQuantity int     `json:"total_quantity" validate:"required,min=0"`
}

type EquipmentRes struct {
	ID            *string   `db:"id" json:"id,omitempty"`
	Name          string    `db:"name" json:"name"`
	CategoryID    string    `db:"category_id" json:"category_id"`
	Ownership     string    `db:"ownership" json:"ownership"`
	IsShared      bool      `db:"is_shared" json:"is_shared"`
	TotalQuantity int       `db:"total_quantity" json:"total_quantity"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

// AssignEquipmentRequest - used when assigning equipment to an employee
type AssignEquipmentRequest struct {
	EmployeeID  uuid.UUID `json:"employee_id" validate:"required"`    // Employee to assign equipment
	EquipmentID uuid.UUID `json:"equipment_id" validate:"required"`   // Equipment being assigned
	Quantity    int       `json:"quantity" validate:"required,min=1"` // Quantity to assign
}
type AssignEquipmentResponse struct {
	EmployeeName  string `json:"employee_name" db:"employee_name"`   // Name of employee
	EmployeeEmail string `json:"employee_email" db:"employee_email"` // Email of employee
	EquipmentName string `json:"equipment_name" db:"equipment_name"` // Equipment name
	Ownership     string `json:"ownership" db:"ownership"`           // Ownership type (COMPANY or SELF)
	Quantity      int    `json:"quantity" db:"quantity"`             // Assigned quantity
}

// RemoveEquipmentRequest - used when removing/returning equipment from an employee
type RemoveEquipmentRequest struct {
	EmployeeID  uuid.UUID `json:"employee_id" validate:"required"`  // Employee to remove equipment from
	EquipmentID uuid.UUID `json:"equipment_id" validate:"required"` // Equipment being removed
}

// UpdateAssignmentRequest - used for both reassigning equipment and updating quantity
type UpdateAssignmentRequest struct {
	FromEmployeeID uuid.UUID  `json:"from_employee_id" validate:"required"` // Current employee
	ToEmployeeID   *uuid.UUID `json:"to_employee_id,omitempty"`             // New employee (optional - if provided, it's reassignment)
	EquipmentID    uuid.UUID  `json:"equipment_id" validate:"required"`     // Equipment being updated/reassigned
	Quantity       int        `json:"quantity" validate:"required,min=1"`   // Quantity to update/reassign
}
