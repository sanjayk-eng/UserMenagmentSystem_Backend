package repositories

import (
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sanjayk-eng/UserMenagmentSystem_Backend/models"
)

// 1. Get leave type entitlement
func (r *Repository) GetLeaveTypeById(tx *sqlx.Tx, leaveTypeID int) (models.LeaveType, error) {
	var leaves models.LeaveType
	query := `SELECT id, name, is_paid, default_entitlement,  created_at, updated_at FROM Tbl_Leave_type WHERE id=$1`
	err := r.DB.Get(&leaves,
		query,
		leaveTypeID,
	)
	return leaves, err
}

func (r *Repository) GetAllLeaveType() ([]models.LeaveType, error) {
	var leaveType []models.LeaveType
	query := `SELECT id, name, is_paid, default_entitlement,  created_at, updated_at FROM Tbl_Leave_type ORDER BY id`
	err := r.DB.Select(&leaveType, query)
	return leaveType, err
}

// 3. Get leave balance (inside TX)
func (r *Repository) GetLeaveBalance(tx *sqlx.Tx, employeeID uuid.UUID, leaveTypeID int) (float64, error) {
	var balance float64
	err := tx.Get(&balance, `
		SELECT closing 
		FROM Tbl_Leave_balance 
		WHERE employee_id=$1 AND leave_type_id=$2 
		AND year = EXTRACT(YEAR FROM CURRENT_DATE)
	`, employeeID, leaveTypeID)
	return balance, err
}

// create leave balance
func (r *Repository) CreateLeaveBalance(tx *sqlx.Tx, employeeID uuid.UUID, leaveTypeID int, entitlement int) error {
	_, err := tx.Exec(`
		INSERT INTO Tbl_Leave_balance 
			(employee_id, leave_type_id, year, opening, accrued, used, adjusted, closing)
		VALUES ($1, $2, EXTRACT(YEAR FROM CURRENT_DATE), $3, 0, 0, 0, $3)
	`, employeeID, leaveTypeID, entitlement)
	return err
}

// 5. Check overlapping leaves
func (r *Repository) GetOverlappingLeaves(
	tx *sqlx.Tx,
	employeeID uuid.UUID,
	startDate, endDate time.Time,
) ([]struct {
	ID        uuid.UUID `db:"id"`
	LeaveType string    `db:"leave_type"`
	StartDate time.Time `db:"start_date"`
	EndDate   time.Time `db:"end_date"`
	Status    string    `db:"status"`
}, error) {

	var result []struct {
		ID        uuid.UUID `db:"id"`
		LeaveType string    `db:"leave_type"`
		StartDate time.Time `db:"start_date"`
		EndDate   time.Time `db:"end_date"`
		Status    string    `db:"status"`
	}

	err := tx.Select(&result, `
		SELECT l.id, lt.name as leave_type, l.start_date, l.end_date, l.status
		FROM Tbl_Leave l
		JOIN Tbl_Leave_type lt ON l.leave_type_id = lt.id
		WHERE l.employee_id=$1 
		AND l.status IN ('Pending','APPROVED')
		AND l.start_date <= $2 
		AND l.end_date >= $3
	`, employeeID, endDate, startDate)

	return result, err
}

func (r *Repository) InsertLeave(
	tx *sqlx.Tx,
	employeeID uuid.UUID,
	leaveTypeID int,
	startDate, endDate time.Time,
	days float64,
	reason string,
) (uuid.UUID, error) {

	var leaveID uuid.UUID

	err := tx.QueryRow(`
		INSERT INTO Tbl_Leave 
		(employee_id, leave_type_id, start_date, end_date, days, status, reason)
		VALUES ($1,$2,$3,$4,$5,'Pending',$6)
		RETURNING id
	`,
		employeeID,
		leaveTypeID,
		startDate,
		endDate,
		days,
		reason,
	).Scan(&leaveID)

	return leaveID, err
}
