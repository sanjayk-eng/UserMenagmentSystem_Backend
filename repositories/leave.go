package repositories

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sanjayk-eng/UserMenagmentSystem_Backend/models"
)

// 1. Get leave type entitlement
func (r *Repository) GetLeaveTypeByIdTx(tx *sqlx.Tx, leaveTypeID int) (models.LeaveType, error) {
	var leaves models.LeaveType
	query := `SELECT id, name, is_paid, default_entitlement,  created_at, updated_at FROM Tbl_Leave_type WHERE id=$1`
	err := tx.Get(&leaves,
		query,
		leaveTypeID,
	)
	return leaves, err
}

// 1. Get leave type entitlement
func (r *Repository) GetLeaveTypeById(leaveTypeID int) (models.LeaveType, error) {
	var leaves models.LeaveType
	query := `SELECT id, name, is_paid, default_entitlement,  created_at, updated_at FROM Tbl_Leave_type WHERE id=$1`
	err := r.DB.Get(&leaves,
		query,
		leaveTypeID,
	)
	return leaves, err
}

func (q *Repository) GetLeaveTypeByLeaveID(leaveID uuid.UUID) (int, error) {
	var leaveTypeID int
	err := q.DB.Get(&leaveTypeID, `
        SELECT leave_type_id 
        FROM Tbl_Leave 
        WHERE id = $1
    `, leaveID)

	if err != nil {
		return 0, err
	}

	return leaveTypeID, nil
}

func (r *Repository) GetAllLeaveType() ([]models.LeaveType, error) {
	var leaveType []models.LeaveType
	query := `SELECT id, name, is_paid, default_entitlement,  created_at, updated_at FROM Tbl_Leave_type ORDER BY id`
	err := r.DB.Select(&leaveType, query)
	return leaveType, err
}

// Admin add leave type

func (r *Repository) AddLeaveType(tx *sqlx.Tx, input models.LeaveTypeInput) (models.LeaveType, error) {
	var leave models.LeaveType
	query := `
		INSERT INTO Tbl_Leave_type (name, is_paid, default_entitlement)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`
	err := tx.QueryRow(query, input.Name, *input.IsPaid, *input.DefaultEntitlement).
		Scan(&leave.ID, &leave.CreatedAt, &leave.UpdatedAt)
	return leave, err
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
	leaveTimingID int,
	startDate, endDate time.Time,
	days float64,
	reason string,
) (uuid.UUID, error) {

	var leaveID uuid.UUID

	err := tx.QueryRow(`
		INSERT INTO Tbl_Leave 
		(employee_id, leave_type_id, half_id, start_date, end_date, days, status, reason)
		VALUES ($1,$2,$3,$4,$5,$6,'Pending',$7)
		RETURNING id
	`,
		employeeID,
		leaveTypeID,
		leaveTimingID,
		startDate,
		endDate,
		days,
		reason,
	).Scan(&leaveID)

	return leaveID, err
}

func (r *Repository) GetLeaveById(tx *sqlx.Tx, leaveID uuid.UUID) (models.Leave, error) {
	var leave models.Leave
	query := `SELECT * FROM Tbl_Leave WHERE id=$1 FOR UPDATE`
	err := tx.Get(&leave, query, leaveID)
	return leave, err
}

// Get All Leave Timming
// Get All Leave Timing
func (r *Repository) GetLeaveTiming() ([]models.LeaveTimingResponse, error) {
	var data []models.LeaveTimingResponse
	query := `
		SELECT id, type, timing, created_at, updated_at
		FROM Tbl_Half
		ORDER BY id
	`
	err := r.DB.Select(&data, query)

	return data, err
}

// Get Leave Timing By ID
func (r *Repository) GetLeaveTimingByID(id int) (*models.LeaveTimingResponse, error) {
	var data models.LeaveTimingResponse

	query := `
		SELECT *
		FROM Tbl_Half
		WHERE id = $1
	`

	err := r.DB.Get(&data, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}
	return &data, nil
}

func (r *Repository) UpdateLeaveTiming(tx *sqlx.Tx, id int, timing string) error {
	query := `
		UPDATE Tbl_Half
		SET timing = $1,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`

	res, err := tx.Exec(query, timing, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// rolde base get leave
func (r *Repository) GetAllEmployeeLeave(userID uuid.UUID) ([]models.LeaveResponse, error) {
	var result []models.LeaveResponse
	query := `
		SELECT 
			l.id,
			e.full_name AS employee,
			lt.name AS leave_type,
			lt.is_paid AS is_paid,
			COALESCE(h.type, 'FULL') AS leave_timing_type,
			COALESCE(h.timing, 'Full Day') AS leave_timing,
			l.start_date,
			l.end_date,
			l.days,
			COALESCE(l.reason, '') AS reason,
			l.status,
			l.created_at AS applied_at
		FROM Tbl_Leave l
		INNER JOIN Tbl_Employee e ON l.employee_id = e.id
		INNER JOIN Tbl_Leave_Type lt ON lt.id = l.leave_type_id
		LEFT JOIN Tbl_Half h ON l.half_id = h.id
		WHERE l.employee_id = $1
		ORDER BY l.created_at DESC`

	err := r.DB.Select(&result, query, userID)
	return result, err
}
func (r *Repository) GetAllleavebaseonassignManager(userID uuid.UUID) ([]models.LeaveResponse, error) {

	var result []models.LeaveResponse
	query := `
		SELECT 
			l.id,
			e.full_name AS employee,
			lt.name AS leave_type,
			lt.is_paid AS is_paid,
			COALESCE(h.type, 'FULL') AS leave_timing_type,
			COALESCE(h.timing, 'Full Day') AS leave_timing,
			l.start_date,
			l.end_date,
			l.days,
			COALESCE(l.reason, '') AS reason,
			l.status,
			l.created_at AS applied_at
		FROM Tbl_Leave l
		INNER JOIN Tbl_Employee e ON l.employee_id = e.id
		INNER JOIN Tbl_Leave_Type lt ON lt.id = l.leave_type_id
		LEFT JOIN Tbl_Half h ON l.half_id = h.id
		WHERE (e.manager_id = $1 OR l.employee_id = $1)
		ORDER BY l.created_at DESC`

	err := r.DB.Select(&result, query, userID)
	return result, err
}

func (r *Repository) GetAllLeave() ([]models.LeaveResponse, error) {
	var result []models.LeaveResponse
	query := `
		SELECT 
			l.id,
			e.full_name AS employee,
			lt.name AS leave_type,
			lt.is_paid AS is_paid,
			COALESCE(h.type, 'FULL') AS leave_timing_type,
			COALESCE(h.timing, 'Full Day') AS leave_timing,
			l.start_date,
			l.end_date,
			l.days,
			COALESCE(l.reason, '') AS reason,
			l.status,
			l.created_at AS applied_at
		FROM Tbl_Leave l
		INNER JOIN Tbl_Employee e ON l.employee_id = e.id
		INNER JOIN Tbl_Leave_Type lt ON lt.id = l.leave_type_id
		LEFT JOIN Tbl_Half h ON l.half_id = h.id
		ORDER BY l.created_at DESC`

	err := r.DB.Select(&result, query)
	return result, err

}

// UpdateLeaveType - Update leave policy
func (r *Repository) UpdateLeaveType(tx *sqlx.Tx, leaveTypeID int, input models.LeaveTypeInput) error {
	query := `
		UPDATE Tbl_Leave_type 
		SET name = $1, is_paid = $2, default_entitlement = $3, updated_at = NOW()
		WHERE id = $4
	`
	result, err := tx.Exec(query, input.Name, *input.IsPaid, *input.DefaultEntitlement, leaveTypeID)
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	
	return nil
}

// DeleteLeaveType - Delete leave policy
func (r *Repository) DeleteLeaveType(tx *sqlx.Tx, leaveTypeID int) error {
	// Check if leave type is being used in any leave applications
	var count int
	err := tx.Get(&count, "SELECT COUNT(*) FROM Tbl_Leave WHERE leave_type_id = $1", leaveTypeID)
	if err != nil {
		return err
	}
	
	if count > 0 {
		return sql.ErrNoRows // Using this to indicate constraint violation
	}
	
	query := `DELETE FROM Tbl_Leave_type WHERE id = $1`
	result, err := tx.Exec(query, leaveTypeID)
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	
	return nil
}
