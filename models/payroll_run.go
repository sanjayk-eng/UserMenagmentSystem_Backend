package models

//
// PayrollRun Model
// -----------------------------------------------------------------------------
// Represents a payroll cycle for a specific month and year.
//
type PayrollRun struct {
	ID     uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Month  int    `json:"month" binding:"required,min=1,max=12"`
	Year   int    `json:"year" binding:"required"`
	Status string `gorm:"type:varchar(20);default:'pending'" json:"status" binding:"oneof=pending finalized"`
}
