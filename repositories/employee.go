package repositories

import "github.com/google/uuid"

// 1. Get employee status
func (r *Repository) GetEmployeeStatus(employeeID uuid.UUID) (string, error) {
	var status string
	err := r.DB.Get(&status, `
		SELECT status FROM Tbl_Employee WHERE id=$1
	`, employeeID)
	return status, err
}
