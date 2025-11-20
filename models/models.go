package models

//
// AllModels
// -----------------------------------------------------------------------------
// Returns all models for automated GORM migration.
//
func AllModels() []interface{} {
	return []interface{}{
		&Employee{},
		&Role{},
		&LeaveType{},
		&Leave{},
		&LeaveBalance{},
		&LeaveAdjustment{},
		&PayrollRun{},
		&Payslip{},
		&Audit{},
	}
}
