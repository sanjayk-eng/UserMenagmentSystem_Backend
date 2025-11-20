package models

//
// Role Model
// -----------------------------------------------------------------------------
// Stores system roles (Admin, Manager, Employee, HR, etc.)
// Used for access-level management.
//
type Role struct {
	ID       uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	RoleName string `gorm:"type:varchar(50);unique;not null" json:"role_name" binding:"required"`
}
