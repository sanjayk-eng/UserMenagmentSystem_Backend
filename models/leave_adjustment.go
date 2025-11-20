package models

import (
	"time"

	"github.com/go-oauth2/oauth2/utils/uuid"
)

// LeaveAdjustment Model
// -----------------------------------------------------------------------------
// HR/Admin adjusts employee leave (manual credit/debit).
type LeaveAdjustment struct {
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`

	EmployeeID uuid.UUID `gorm:"type:uuid;not null" json:"employee_id" binding:"required"`
	Employee   Employee  `gorm:"foreignKey:EmployeeID"`

	LeaveTypeID uint      `gorm:"not null" json:"leave_type_id" binding:"required"`
	LeaveType   LeaveType `gorm:"foreignKey:LeaveTypeID"`

	Quantity  int       `json:"quantity" binding:"required"`
	Reason    string    `gorm:"type:varchar(255)" json:"reason" binding:"required"`
	CreatedBy uuid.UUID `gorm:"type:uuid;not null" json:"created_by" binding:"required"`

	CreatedAt time.Time `json:"created_at"`
}
