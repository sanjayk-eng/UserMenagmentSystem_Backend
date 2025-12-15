package repositories

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sanjayk-eng/UserMenagmentSystem_Backend/models"
)

// ======================
// Category Repositories
// ======================

// CreateCatagory inserts a new equipment category into the database.
// Uses transaction tx to allow rollback if needed.
func (r *Repository) CreateCatagory(tx *sqlx.Tx, data models.EquipmentCategoryRequest) error {
	query := `
		INSERT INTO tbl_equipment_category (name, description)
		VALUES ($1, $2)
	`

	_, err := tx.Exec(query, data.Name, data.Description)
	if err != nil {
		return fmt.Errorf("failed to create category: %w", err)
	}
	return nil
}

// GetAllCategory fetches all equipment categories from the database.
// Returns a slice of EquipmentCategoryRes or an error.
func (r *Repository) GetAllCategory() ([]models.EquipmentCategoryRes, error) {
	var categories []models.EquipmentCategoryRes

	query := `
		SELECT *
		FROM tbl_equipment_category
		ORDER BY name ASC
	`

	err := r.DB.Select(&categories, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch categories: %w", err)
	}
	return categories, nil
}

// DeleteCategory deletes a category by its UUID.
// Checks rows affected to ensure the category exists.
func (r *Repository) DeleteCategory(id uuid.UUID) error {
	query := `
		DELETE FROM tbl_equipment_category
		WHERE id = $1
	`

	result, err := r.DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	// Ensure at least one row was deleted
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("category with id %s not found", id)
	}
	return nil
}

// UpdateCategory updates an existing category by UUID.
// Uses transaction tx to allow rollback and updates the updated_at timestamp.
func (r *Repository) UpdateCategory(tx *sqlx.Tx, id uuid.UUID, data models.EquipmentCategoryRequest) error {
	query := `
		UPDATE tbl_equipment_category
		SET name = $1,
		    description = $2,
		    updated_at = now()
		WHERE id = $3
	`

	result, err := tx.Exec(query, data.Name, data.Description, id)
	if err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}

	// Ensure at least one row was updated
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("category with id %s not found", id)
	}
	return nil
}

// ======================
// Equipment Repositories
// ======================

// CreateEquipment inserts a new equipment record into the database.
// Uses transaction tx for rollback if needed.
func (r *Repository) CreateEquipment(tx *sqlx.Tx, data models.EquipmentRequest) error {
	query := `
		INSERT INTO tbl_equipment 
			(name, category_id, ownership, is_shared, total_quantity)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := tx.Exec(query, data.Name, data.CategoryID, data.Ownership, data.IsShared, data.TotalQuantity)
	if err != nil {
		return fmt.Errorf("failed to create equipment: %w", err)
	}
	return nil
}

// GetEquipmentByCategory fetches equipment filtered by category ID
func (r *Repository) GetEquipmentByCategory(categoryID uuid.UUID) ([]models.EquipmentRes, error) {
	// Initialize with empty slice to ensure JSON returns [] instead of null
	equipments := make([]models.EquipmentRes, 0)

	query := `
		SELECT *
		FROM tbl_equipment
		WHERE category_id = $1
		ORDER BY name ASC
	`

	err := r.DB.Select(&equipments, query, categoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch equipment: %w", err)
	}

	// Return empty array instead of error when no equipment found
	return equipments, nil
}

// GetAllEquipment fetches all equipment records from the database.
// Returns a slice of EquipmentRes or an empty array if none found.
func (r *Repository) GetAllEquipment() ([]models.EquipmentRes, error) {
	// Initialize with empty slice to ensure JSON returns [] instead of null
	equipments := make([]models.EquipmentRes, 0)

	query := `
		SELECT *
		FROM tbl_equipment
		ORDER BY name ASC
	`

	err := r.DB.Select(&equipments, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch equipment: %w", err)
	}

	// Return empty array instead of error when no equipment found
	return equipments, nil
}

// UpdateEquipment updates an existing equipment by UUID.
// Uses transaction tx and updates updated_at timestamp.
func (r *Repository) UpdateEquipment(tx *sqlx.Tx, id uuid.UUID, data models.EquipmentRequest) error {
	query := `
		UPDATE tbl_equipment
		SET name = $1,
		    category_id = $2,
		    ownership = $3,
		    is_shared = $4,
		    total_quantity = $5,
		    updated_at = now()
		WHERE id = $6
	`

	result, err := tx.Exec(query, data.Name, data.CategoryID, data.Ownership, data.IsShared, data.TotalQuantity, id)
	if err != nil {
		return fmt.Errorf("failed to update equipment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("equipment with id %s not found", id)
	}

	return nil
}

// DeleteEquipment deletes an equipment record by UUID.
// Checks rows affected to ensure it exists.
func (r *Repository) DeleteEquipment(tx *sqlx.Tx, id uuid.UUID) error {
	query := `
		DELETE FROM tbl_equipment
		WHERE id = $1
	`

	result, err := tx.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete equipment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("equipment with id %s not found", id)
	}

	return nil
}

func (r *Repository) AssignEquipment(tx *sqlx.Tx, req models.AssignEquipmentRequest) error {
	// 1. Check if enough quantity is available
	var totalQuantity int
	err := tx.Get(&totalQuantity, "SELECT total_quantity FROM tbl_equipment WHERE id=$1", req.EquipmentID)
	if err != nil {
		return fmt.Errorf("equipment not found")
	}

	var assignedQuantity int
	err = tx.Get(&assignedQuantity, "SELECT COALESCE(SUM(quantity),0) FROM tbl_equipment_assignment WHERE equipment_id=$1 AND returned_at IS NULL", req.EquipmentID)
	if err != nil {
		return fmt.Errorf("failed to get assigned quantity")
	}

	if totalQuantity-assignedQuantity < req.Quantity {
		return fmt.Errorf("not enough equipment quantity available")
	}

	// 2. Check if employee already has this equipment assigned
	var exists int
	err = tx.Get(&exists, "SELECT COUNT(*) FROM tbl_equipment_assignment WHERE equipment_id=$1 AND employee_id=$2 AND returned_at IS NULL", req.EquipmentID, req.EmployeeID)
	if err != nil {
		return fmt.Errorf("failed to check existing assignment")
	}
	if exists > 0 {
		return fmt.Errorf("equipment already assigned to this employee")
	}

	// 3. Insert assignment
	_, err = tx.Exec("INSERT INTO tbl_equipment_assignment (equipment_id, employee_id, quantity) VALUES ($1,$2,$3)", req.EquipmentID, req.EmployeeID, req.Quantity)
	if err != nil {
		return fmt.Errorf("failed to assign equipment: %v", err)
	}

	return nil
}

// Get all assigned equipment
func (r *Repository) GetAllAssignedEquipment() ([]models.AssignEquipmentResponse, error) {
	// Initialize with empty slice to ensure JSON returns [] instead of null
	result := make([]models.AssignEquipmentResponse, 0)
	query := `
	SELECT e.full_name as employee_name, e.email as employee_email,
	       eq.name as equipment_name, eq.ownership, ea.quantity
	FROM tbl_equipment_assignment ea
	JOIN tbl_employee e ON ea.employee_id = e.id
	JOIN tbl_equipment eq ON ea.equipment_id = eq.id
	WHERE ea.returned_at IS NULL
	ORDER BY ea.assigned_at DESC
	`
	err := r.DB.Select(&result, query)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Get assigned equipment by employee ID
func (r *Repository) GetAssignedEquipmentByEmployee(employeeID string) ([]models.AssignEquipmentResponse, error) {
	// Initialize with empty slice to ensure JSON returns [] instead of null
	result := make([]models.AssignEquipmentResponse, 0)
	query := `
	SELECT e.full_name as employee_name, e.email as employee_email,
	       eq.name as equipment_name, eq.ownership, ea.quantity
	FROM tbl_equipment_assignment ea
	JOIN tbl_employee e ON ea.employee_id = e.id
	JOIN tbl_equipment eq ON ea.equipment_id = eq.id
	WHERE ea.returned_at IS NULL AND e.id=$1
	ORDER BY ea.assigned_at DESC
	`
	err := r.DB.Select(&result, query, employeeID)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// RemoveEquipment removes/returns equipment from an employee
func (r *Repository) RemoveEquipment(tx *sqlx.Tx, req models.RemoveEquipmentRequest) error {
	// Check if assignment exists
	var exists int
	err := tx.Get(&exists, "SELECT COUNT(*) FROM tbl_equipment_assignment WHERE equipment_id=$1 AND employee_id=$2 AND returned_at IS NULL", req.EquipmentID, req.EmployeeID)
	if err != nil {
		return fmt.Errorf("failed to check assignment: %v", err)
	}
	if exists == 0 {
		return fmt.Errorf("no active assignment found for this equipment and employee")
	}

	// Mark as returned
	_, err = tx.Exec("UPDATE tbl_equipment_assignment SET returned_at = NOW() WHERE equipment_id=$1 AND employee_id=$2 AND returned_at IS NULL", req.EquipmentID, req.EmployeeID)
	if err != nil {
		return fmt.Errorf("failed to remove equipment: %v", err)
	}

	return nil
}

// UpdateAssignment handles both quantity updates and reassignments
func (r *Repository) UpdateAssignment(tx *sqlx.Tx, req models.UpdateAssignmentRequest) error {
	// 1. Check if source assignment exists
	var currentQuantity int
	err := tx.Get(&currentQuantity, "SELECT quantity FROM tbl_equipment_assignment WHERE equipment_id=$1 AND employee_id=$2 AND returned_at IS NULL", req.EquipmentID, req.FromEmployeeID)
	if err != nil {
		return fmt.Errorf("no active assignment found for source employee")
	}

	// 2. If ToEmployeeID is provided, it's a reassignment
	if req.ToEmployeeID != nil {
		// REASSIGNMENT LOGIC
		if currentQuantity < req.Quantity {
			return fmt.Errorf("requested quantity exceeds assigned quantity")
		}

		// Check if target employee already has this equipment
		var targetExists int
		err = tx.Get(&targetExists, "SELECT COUNT(*) FROM tbl_equipment_assignment WHERE equipment_id=$1 AND employee_id=$2 AND returned_at IS NULL", req.EquipmentID, *req.ToEmployeeID)
		if err != nil {
			return fmt.Errorf("failed to check target assignment")
		}
		if targetExists > 0 {
			return fmt.Errorf("target employee already has this equipment assigned")
		}

		// Update source assignment quantity or mark as returned
		if currentQuantity == req.Quantity {
			// Mark as returned if reassigning all quantity
			_, err = tx.Exec("UPDATE tbl_equipment_assignment SET returned_at = NOW() WHERE equipment_id=$1 AND employee_id=$2 AND returned_at IS NULL", req.EquipmentID, req.FromEmployeeID)
		} else {
			// Reduce quantity
			_, err = tx.Exec("UPDATE tbl_equipment_assignment SET quantity = quantity - $1 WHERE equipment_id=$2 AND employee_id=$3 AND returned_at IS NULL", req.Quantity, req.EquipmentID, req.FromEmployeeID)
		}
		if err != nil {
			return fmt.Errorf("failed to update source assignment: %v", err)
		}

		// Create new assignment for target employee
		_, err = tx.Exec("INSERT INTO tbl_equipment_assignment (equipment_id, employee_id, quantity) VALUES ($1,$2,$3)", req.EquipmentID, *req.ToEmployeeID, req.Quantity)
		if err != nil {
			return fmt.Errorf("failed to create new assignment: %v", err)
		}

	} else {
		// QUANTITY UPDATE LOGIC (same employee)
		// If increasing quantity, check availability
		if req.Quantity > currentQuantity {
			var totalQuantity int
			err := tx.Get(&totalQuantity, "SELECT total_quantity FROM tbl_equipment WHERE id=$1", req.EquipmentID)
			if err != nil {
				return fmt.Errorf("equipment not found")
			}

			var assignedQuantity int
			err = tx.Get(&assignedQuantity, "SELECT COALESCE(SUM(quantity),0) FROM tbl_equipment_assignment WHERE equipment_id=$1 AND returned_at IS NULL", req.EquipmentID)
			if err != nil {
				return fmt.Errorf("failed to get assigned quantity")
			}

			// Calculate available quantity (subtract current assignment from total assigned)
			availableQuantity := totalQuantity - (assignedQuantity - currentQuantity)
			if availableQuantity < req.Quantity {
				return fmt.Errorf("not enough equipment quantity available")
			}
		}

		// Update assignment quantity for same employee
		_, err = tx.Exec("UPDATE tbl_equipment_assignment SET quantity = $1 WHERE equipment_id=$2 AND employee_id=$3 AND returned_at IS NULL", req.Quantity, req.EquipmentID, req.FromEmployeeID)
		if err != nil {
			return fmt.Errorf("failed to update assignment: %v", err)
		}
	}

	return nil
}
