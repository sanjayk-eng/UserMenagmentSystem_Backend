package models

import (
	"time"

	"github.com/go-oauth2/oauth2/utils/uuid"
)

// Employee Model
// -----------------------------------------------------------------------------
// Represents a company employee.
// Includes authentication fields, HR data, reporting hierarchy,
// soft-deletion support, and relational mapping with Role and Manager.
type Employee struct {
	// Primary Key (UUID)
	ID uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`

	// Basic Employee Information
	FullName string `gorm:"type:varchar(100);not null" json:"full_name" binding:"required"`
	Email    string `gorm:"type:varchar(100);uniqueIndex;not null" json:"email" binding:"required,email,endswith=zenithive.com"`
	Password string `gorm:"not null" json:"password" binding:"required,min=6"`

	// Role Mapping (Foreign Key)
	RoleID uint `gorm:"not null" json:"role_id" binding:"required"`
	Role   Role `gorm:"foreignKey:RoleID"`

	// Manager -> Employee Self Relationship
	ManagerID *uuid.UUID `gorm:"type:uuid" json:"manager_id"`
	Manager   *Employee  `gorm:"foreignKey:ManagerID"`

	// HR Information
	Salary      float64   `gorm:"not null" json:"salary" binding:"required,gt=0"`
	JoiningDate time.Time `gorm:"not null" json:"joining_date" binding:"required"`

	// active | inactive | suspended
	Status string `gorm:"type:varchar(20);default:'active'" json:"status" binding:"oneof=active inactive suspended"`

	// Soft Delete Support
	Deleted bool `gorm:"default:false" json:"deleted"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
