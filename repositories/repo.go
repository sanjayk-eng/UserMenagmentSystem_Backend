package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sanjayk-eng/UserMenagmentSystem_Backend/models"
)

type EmployeeAuthData struct {
	ID       string `db:"id"`
	Email    string `db:"email"`
	Password string `db:"password"`
	Role     string `db:"role"`
	Status   string `db:"status"`
}

type Repository struct {
	DB *sqlx.DB
}

func InitializeRepo(db *sqlx.DB) *Repository {
	return &Repository{
		DB: db,
	}
}
func (r *Repository) GetEmployeeByEmail(email string) (EmployeeAuthData, error) {
	var emp EmployeeAuthData

	query := `
		SELECT 
			e.id,
			e.email,
			e.password,
			r.type AS role,
			e.status
		FROM Tbl_Employee e
		JOIN Tbl_Role r ON e.role_id = r.id
		WHERE e.email = $1
		LIMIT 1;
	`

	err := r.DB.Get(&emp, query, email)
	return emp, err
}

func (r *Repository) GetAllEmployees() ([]models.EmployeeInput, error) {
	query := `
        SELECT 
            e.id, e.full_name, e.email, e.status,
            r.type AS role, e.password, e.manager_id,
            e.salary, e.joining_date,
            e.created_at, e.updated_at, e.deleted_at
        FROM Tbl_Employee e
        JOIN Tbl_Role r ON e.role_id = r.id
        ORDER BY e.full_name
    `

	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var employees []models.EmployeeInput

	for rows.Next() {
		var emp models.EmployeeInput

		err := rows.Scan(
			&emp.ID,
			&emp.FullName,
			&emp.Email,
			&emp.Status,
			&emp.Role,
			&emp.Password,
			&emp.ManagerID,
			&emp.Salary,
			&emp.JoiningDate,
			&emp.CreatedAt,
			&emp.UpdatedAt,
			&emp.DeletedAt,
		)
		if err != nil {
			return nil, err
		}

		// ------- Fetch manager name if exists -------
		if emp.ManagerID != nil {
			var mName string
			err := r.DB.QueryRow(`
                SELECT full_name FROM Tbl_Employee WHERE id = $1
            `, emp.ManagerID).Scan(&mName)

			if err == nil {
				emp.ManagerName = &mName
			}
		}

		employees = append(employees, emp)
	}

	return employees, nil
}

func (r *Repository) DeleteEmployeeStatus(id uuid.UUID) (string, error) {

	// Get current status
	var currentStatus string
	err := r.DB.QueryRow(`
        SELECT status FROM Tbl_Employee WHERE id = $1
    `, id).Scan(&currentStatus)
	if err != nil {
		return "", err
	}

	// Toggle logic
	newStatus := "active"
	if currentStatus == "active" {
		newStatus = "deactive"
	}

	// Update
	_, err = r.DB.Exec(`
        UPDATE Tbl_Employee 
        SET status = $1, updated_at = NOW()
        WHERE id = $2
    `, newStatus, id)
	if err != nil {
		return "", err
	}

	return newStatus, nil
}

// ------------------ CHECK EMAIL EXISTS ------------------
func (r *Repository) CheckEmailExists(email string) (bool, error) {
	var existing string
	err := r.DB.QueryRow(
		`SELECT email FROM Tbl_Employee WHERE email=$1`, email,
	).Scan(&existing)

	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

// ------------------ GET ROLE ID ------------------
func (r *Repository) GetRoleID(role string) (string, error) {
	var id string
	err := r.DB.QueryRow(`SELECT id FROM Tbl_Role WHERE type=$1`, role).Scan(&id)
	return id, err
}

// ------------------ CREATE EMPLOYEE ------------------
func (r *Repository) InsertEmployee(fullName, email, roleID, password string, salary *float64, joining *time.Time) error {
	_, err := r.DB.Exec(`
		INSERT INTO Tbl_Employee (full_name, email, role_id, password, salary, joining_date)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, fullName, email, roleID, password, salary, joining)
	return err
}

// ------------------ GET CURRENT ROLE NAME ------------------
func (r *Repository) GetEmployeeCurrentRole(empID string) (string, error) {
	var role string
	err := r.DB.QueryRow(`
        SELECT R.TYPE
        FROM TBL_EMPLOYEE E
        JOIN TBL_ROLE R ON E.ROLE_ID = R.ID
        WHERE E.ID = $1
    `, empID).Scan(&role)
	return role, err
}

// ------------------ UPDATE ROLE ------------------
func (r *Repository) UpdateEmployeeRole(empID uuid.UUID, newRole string) (string, error) {
	var id string
	query := `
        UPDATE TBL_EMPLOYEE
        SET ROLE_ID = (SELECT ID FROM TBL_ROLE WHERE TYPE=$1),
            UPDATED_AT = NOW()
        WHERE ID = $2
        RETURNING ID;
    `
	err := r.DB.QueryRow(query, newRole, empID).Scan(&id)
	return id, err
}

// ------------------ CHECK MANAGER EXISTS ------------------
func (r *Repository) ManagerExists(id uuid.UUID) (bool, error) {
	var exists bool
	err := r.DB.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM TBL_EMPLOYEE WHERE ID=$1)`,
		id,
	).Scan(&exists)
	return exists, err
}

// ------------------ UPDATE MANAGER ------------------
func (r *Repository) UpdateManager(empID, managerID uuid.UUID) error {
	_, err := r.DB.Exec(`
        UPDATE TBL_EMPLOYEE
        SET MANAGER_ID=$1, UPDATED_AT=NOW()
        WHERE ID=$2
    `, managerID, empID)
	return err
}

// AddHoliday inserts a holiday into the database
func (r *Repository) AddHoliday(name string, date time.Time, typ string) (string, error) {
	if typ == "" {
		typ = "HOLIDAY"
	}
	day := date.Weekday().String()
	var id string
	err := r.DB.QueryRow(`
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
func (r *Repository) DeleteHoliday(id string) error {
	_, err := r.DB.Exec(`DELETE FROM Tbl_Holiday WHERE id=$1`, id)
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

func (r *Repository) GetEmployeeCurrentRoleAndManagerStatus(empID uuid.UUID) (string, bool, error) {
	var role string
	var count int
	query := `
	SELECT r.type, 
	       (SELECT COUNT(*) FROM Tbl_Employee e2 WHERE e2.manager_id=e.id) AS sub_count
	FROM Tbl_Employee e
	JOIN Tbl_Role r ON e.role_id=r.id
	WHERE e.id=$1
	`
	err := r.DB.QueryRow(query, empID).Scan(&role, &count)
	if err != nil {
		return "", false, err
	}
	return role, count > 0, nil
}

func (r *Repository) GetAllFinalizedPayslips() (*sql.Rows, error) {
	query := `
	SELECT 
	    p.id AS payslip_id,
	    e.id AS employee_id,
	    e.full_name,
	    e.email,
	    pr.month,
	    pr.year,
	    p.basic_salary,
	    p.working_days,
	    p.absent_days,
	    p.deduction_amount,
	    p.net_salary,
	    COALESCE(p.pdf_path, '') AS pdf_path,
	    CONCAT('₹', p.basic_salary, ' - ₹', p.deduction_amount, ' = ₹', p.net_salary) AS calculation,
	    p.created_at
	FROM Tbl_Payslip p
	JOIN Tbl_Employee e ON p.employee_id = e.id
	JOIN Tbl_Payroll_Run pr ON pr.id = p.payroll_run_id
	WHERE pr.status = 'FINALIZED'
	ORDER BY pr.year DESC, pr.month DESC, e.full_name ASC;
	`
	return r.DB.Query(query)
}

func (r *Repository) GetFinalizedPayslipsByEmployee(id uuid.UUID) (*sql.Rows, error) {
	query := `
	SELECT 
	    p.id AS payslip_id,
	    e.id AS employee_id,
	    e.full_name,
	    e.email,
	    pr.month,
	    pr.year,
	    p.basic_salary,
	    p.working_days,
	    p.absent_days,
	    p.deduction_amount,
	    p.net_salary,
	    COALESCE(p.pdf_path, '') AS pdf_path,
	    CONCAT('₹', p.basic_salary, ' - ₹', p.deduction_amount, ' = ₹', p.net_salary) AS calculation,
	    p.created_at
	FROM Tbl_Payslip p
	JOIN Tbl_Employee e ON p.employee_id = e.id
	JOIN Tbl_Payroll_Run pr ON pr.id = p.payroll_run_id
	WHERE pr.status = 'FINALIZED' AND e.id = $1
	ORDER BY pr.year DESC, pr.month DESC;
	`
	return r.DB.Query(query, id)
}

// ------------------ GET EMPLOYEE BY ID ------------------
func (r *Repository) GetEmployeeByID(empID uuid.UUID) (*models.EmployeeInput, error) {
	var emp models.EmployeeInput
	query := `
        SELECT 
            e.id, e.full_name, e.email, e.status,
            r.type AS role, e.manager_id,
            e.salary, e.joining_date,
            e.created_at, e.updated_at, e.deleted_at
        FROM Tbl_Employee e
        JOIN Tbl_Role r ON e.role_id = r.id
        WHERE e.id = $1
    `
	
	err := r.DB.QueryRow(query, empID).Scan(
		&emp.ID,
		&emp.FullName,
		&emp.Email,
		&emp.Status,
		&emp.Role,
		&emp.ManagerID,
		&emp.Salary,
		&emp.JoiningDate,
		&emp.CreatedAt,
		&emp.UpdatedAt,
		&emp.DeletedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	// Fetch manager name if exists
	if emp.ManagerID != nil {
		var mName string
		err := r.DB.QueryRow(`
            SELECT full_name FROM Tbl_Employee WHERE id = $1
        `, emp.ManagerID).Scan(&mName)
		
		if err == nil {
			emp.ManagerName = &mName
		}
	}
	
	return &emp, nil
}

// ------------------ UPDATE EMPLOYEE INFO ------------------
func (r *Repository) UpdateEmployeeInfo(empID uuid.UUID, fullName, email string, salary *float64) error {
	_, err := r.DB.Exec(`
        UPDATE Tbl_Employee
        SET full_name = $1, email = $2, salary = $3, updated_at = NOW()
        WHERE id = $4
    `, fullName, email, salary, empID)
	return err
}

// ------------------ UPDATE EMPLOYEE PASSWORD ------------------
func (r *Repository) UpdateEmployeePassword(empID uuid.UUID, hashedPassword string) error {
	_, err := r.DB.Exec(`
        UPDATE Tbl_Employee
        SET password = $1, updated_at = NOW()
        WHERE id = $2
    `, hashedPassword, empID)
	return err
}
