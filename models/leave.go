package models

import (
	"time"

	"github.com/go-oauth2/oauth2/utils/uuid"
)

// Leave Model
// -----------------------------------------------------------------------------
// Stores leave applications of employees.
// Includes leave duration, status, type, and approval workflow.
type Leave struct {
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`

	// Employee Applying Leave
	EmployeeID uuid.UUID `gorm:"type:uuid;not null" json:"employee_id" binding:"required"`
	Employee   Employee  `gorm:"foreignKey:EmployeeID"`

	// Type of Leave (Sick, Casual, etc.)
	LeaveTypeID uint      `gorm:"not null" json:"leave_type_id" binding:"required"`
	LeaveType   LeaveType `gorm:"foreignKey:LeaveTypeID"`

	// Leave Duration
	StartDate time.Time `json:"start_date" binding:"required"`
	EndDate   time.Time `json:"end_date" binding:"required,gtefield=StartDate"`
	Days      int       `json:"days" binding:"required,min=1"`

	// pending | approved | rejected
	Status string `gorm:"type:varchar(20);default:'pending'" json:"status" binding:"oneof=pending approved rejected"`

	// Workflow Users
	AppliedBy  uuid.UUID  `gorm:"type:uuid;not null" json:"applied_by" binding:"required"`
	ApprovedBy *uuid.UUID `gorm:"type:uuid" json:"approved_by"`

	CreatedAt time.Time `json:"created_at"`
}
