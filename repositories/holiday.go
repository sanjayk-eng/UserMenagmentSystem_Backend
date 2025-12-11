package repositories

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sanjayk-eng/UserMenagmentSystem_Backend/models"
)

// AddHoliday inserts a holiday into the database
func (r *Repository) AddHoliday(tx *sqlx.Tx, name string, date time.Time, typ string) (string, error) {
	if typ == "" {
		typ = "HOLIDAY"
	}
	day := date.Weekday().String()
	var id string
	err := tx.QueryRow(`
		INSERT INTO Tbl_Holiday (name, date, day, type, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id
	`, name, date, day, typ).Scan(&id)
	return id, err
}

// GetAllHolidays fetches all holidays
func (r *Repository) GetAllHolidays() ([]models.Holiday, error) {
	rows, err := r.DB.Queryx(`SELECT id, name, date, day, type, created_at, updated_at FROM Tbl_Holiday ORDER BY date`)
	if err != nil {
		fmt.Println("error", err)
		return nil, err
	}
	defer rows.Close()

	var holidays []models.Holiday
	for rows.Next() {
		var h models.Holiday
		if err := rows.StructScan(&h); err != nil {
			return nil, err
		}
		holidays = append(holidays, h)
	}
	return holidays, nil
}

// DeleteHoliday deletes a holiday by ID
func (r *Repository) DeleteHoliday(id string, tx *sqlx.Tx) error {
	_, err := tx.Exec(`DELETE FROM Tbl_Holiday WHERE id=$1`, id)
	return err
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
