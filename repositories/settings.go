package repositories

import (
	"github.com/jmoiron/sqlx"
	"github.com/sanjayk-eng/UserMenagmentSystem_Backend/models"
)

func (r *Repository) GetCompanySettings(settings *models.CompanySettings) error {

	err := r.DB.Get(settings, `SELECT * FROM Tbl_Company_Settings LIMIT 1`)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) UpdateCompanySettings(tx *sqlx.Tx, input models.CompanyField) error {
	_, err := tx.Exec(`
        UPDATE Tbl_Company_Settings
        SET working_days_per_month=$1, allow_manager_add_leave=$2, updated_at=NOW()
    `, input.WorkingDaysPerMonth, input.AllowManagerAddLeave)

	if err != nil {
		return err
	}
	return nil

}
