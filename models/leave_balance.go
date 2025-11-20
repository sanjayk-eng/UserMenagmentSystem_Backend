package models

import "github.com/go-oauth2/oauth2/utils/uuid"

//
// LeaveBalance Model
// -----------------------------------------------------------------------------
// Tracks yearly leave availability for each employee & type.
//
type LeaveBalance struct {
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`

	// User for whom balance is maintained
	EmployeeID uuid.UUID `gorm:"type:uuid;not null" json:"employee_id" binding:"required"`
	Employee   Employee  `gorm:"foreignKey:EmployeeID"`

	// Leave Type
	LeaveTypeID uint      `gorm:"not null" json:"leave_type_id" binding:"required"`
	LeaveType   LeaveType `gorm:"foreignKey:LeaveTypeID"`

	// Yearly Stats
	Year     int `json:"year" binding:"required"`
	Opening  int `json:"opening"`
	Accrued  int `json:"accrued"`
	Used     int `json:"used"`
	Adjusted int `json:"adjusted"`
	Closing  int `json:"closing"`
}
