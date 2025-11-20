package models

//
// LeaveType Model
// -----------------------------------------------------------------------------
// Defines leave categories like Sick Leave, Casual Leave, Paid Leave, etc.
// Default entitlement refers to annually allowed leave quota.
//
type LeaveType struct {
	ID                 uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name               string `gorm:"type:varchar(50);unique;not null" json:"name" binding:"required"`
	IsPaid             bool   `json:"is_paid"`
	DefaultEntitlement int    `json:"default_entitlement" binding:"required,min=0"`
}
