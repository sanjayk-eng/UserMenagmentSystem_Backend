# Designation Feature Documentation

## Overview
This document outlines all changes and additions made to implement the Designation management feature in the User Management System.

---

## Database Changes

### Migration File
**File:** `pkg/migration/20251204073026_add_company_designation_responsibility.sql`

#### New Table: `Tbl_Designation`
```sql
CREATE TABLE IF NOT EXISTS Tbl_Designation (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    designation_name VARCHAR(100) NOT NULL,
    description TEXT
);
```

#### Employee Table Modification
```sql
-- Add designation_id column to Tbl_Employee
ALTER TABLE Tbl_Employee ADD COLUMN designation_id UUID;

-- Add foreign key constraint with ON DELETE SET NULL
ALTER TABLE Tbl_Employee 
ADD CONSTRAINT fk_employee_designation 
FOREIGN KEY (designation_id) 
REFERENCES Tbl_Designation(id) 
ON DELETE SET NULL;
```

**Key Feature:** When a designation is deleted, all employees with that designation will have their `designation_id` automatically set to `NULL` (no data loss).

---

## API Endpoints

### 1. Create Designation
**Endpoint:** `POST /api/designations`

**Access:** ADMIN, SUPERADMIN, HR only

**Request Body:**
```json
{
  "designation_name": "Senior Developer",
  "description": "Senior software developer position"
}
```

**Response:**
```json
{
  "message": "designation created successfully",
  "designation_id": "uuid-here"
}
```

---

### 2. Get All Designations
**Endpoint:** `GET /api/designations`

**Access:** All authenticated users

**Response:**
```json
{
  "message": "designations fetched successfully",
  "designations": [
    {
      "id": "uuid-here",
      "designation_name": "Senior Developer",
      "description": "Senior software developer position"
    },
    {
      "id": "uuid-here",
      "designation_name": "Team Lead",
      "description": "Team leadership role"
    }
  ]
}
```

---

### 3. Get Designation by ID
**Endpoint:** `GET /api/designations/:id`

**Access:** All authenticated users

**Response:**
```json
{
  "message": "designation fetched successfully",
  "designation": {
    "id": "uuid-here",
    "designation_name": "Senior Developer",
    "description": "Senior software developer position"
  }
}
```

---

### 4. Update Designation
**Endpoint:** `PATCH /api/designations/:id`

**Access:** ADMIN, SUPERADMIN, HR only

**Request Body:**
```json
{
  "designation_name": "Lead Developer",
  "description": "Updated description"
}
```

**Response:**
```json
{
  "message": "designation updated successfully",
  "designation_id": "uuid-here"
}
```

---

### 5. Delete Designation
**Endpoint:** `DELETE /api/designations/:id`

**Access:** ADMIN, SUPERADMIN, HR only

**Response:**
```json
{
  "message": "designation deleted successfully. Employee designation_id set to NULL."
}
```

---

### 6. Assign/Update Employee Designation
**Endpoint:** `PATCH /api/employee/:id/designation`

**Access:** ADMIN, SUPERADMIN, HR only

**Request Body:**
```json
{
  "designation_id": "uuid-here"
}
```

**To Remove Designation:**
```json
{
  "designation_id": null
}
```

**Response:**
```json
{
  "message": "employee designation updated successfully",
  "employee_id": "uuid-here",
  "designation_id": "uuid-here"
}
```

**Protection Rules:**
- HR and ADMIN cannot modify SUPERADMIN users' designations
- Designation must exist before assignment
- Employee must exist

---

## Code Changes

### 1. New Controller File
**File:** `controllers/designation.go`

**Functions Added:**
- `CreateDesignation(c *gin.Context)` - Create new designation
- `GetAllDesignations(c *gin.Context)` - Fetch all designations
- `GetDesignationByID(c *gin.Context)` - Fetch single designation
- `UpdateDesignation(c *gin.Context)` - Update designation
- `DeleteDesignation(c *gin.Context)` - Delete designation

**Struct Added:**
```go
type DesignationInput struct {
    DesignationName string  `json:"designation_name" validate:"required"`
    Description     *string `json:"description,omitempty"`
}
```

---

### 2. Employee Controller Updates
**File:** `controllers/employee.go`

**New Function Added:**
```go
func (h *HandlerFunc) UpdateEmployeeDesignation(c *gin.Context)
```

**Features:**
- Assign designation to employee
- Remove designation from employee
- Validate designation exists
- Protect SUPERADMIN from HR/ADMIN modifications

---

### 3. Models Updates
**File:** `models/models.go`

**EmployeeInput Struct - Fields Added:**
```go
DesignationID   *uuid.UUID `json:"designation_id,omitempty"`
DesignationName *string    `json:"designation_name,omitempty"`
```

**New Structs Added:**
```go
type Designation struct {
    ID              string  `json:"id" db:"id"`
    DesignationName string  `json:"designation_name" db:"designation_name"`
    Description     *string `json:"description,omitempty" db:"description"`
}

type DesignationInput struct {
    DesignationName string  `json:"designation_name" validate:"required"`
    Description     *string `json:"description,omitempty"`
}
```

---

### 4. Repository Updates
**File:** `repositories/repo.go`

**New Struct Added:**
```go
type Designation struct {
    ID              uuid.UUID `json:"id" db:"id"`
    DesignationName string    `json:"designation_name" db:"designation_name"`
    Description     *string   `json:"description,omitempty" db:"description"`
}
```

**New Functions Added:**
```go
func (r *Repository) CreateDesignation(name string, description *string) (string, error)
func (r *Repository) GetAllDesignations() ([]Designation, error)
func (r *Repository) GetDesignationByID(id uuid.UUID) (*Designation, error)
func (r *Repository) UpdateDesignation(id uuid.UUID, name string, description *string) error
func (r *Repository) DeleteDesignation(id uuid.UUID) error
func (r *Repository) UpdateEmployeeDesignation(empID uuid.UUID, designationID *uuid.UUID) error
```

**Modified Functions:**
- `GetAllEmployees()` - Now fetches `designation_id` and `designation_name`
- `GetEmployeeByID()` - Now fetches `designation_id` and `designation_name`
- `GetEmployeesByManagerID()` - Now fetches `designation_id` and `designation_name`

**Query Updates:**
All employee queries now include:
```sql
SELECT e.designation_id, ...
```

And fetch designation name:
```go
if emp.DesignationID != nil {
    var dName string
    err := r.DB.QueryRow(`
        SELECT designation_name FROM Tbl_Designation WHERE id = $1
    `, emp.DesignationID).Scan(&dName)
    
    if err == nil {
        emp.DesignationName = &dName
    }
}
```

---

### 5. Routes Updates
**File:** `routes/router.go`

**New Route Group Added:**
```go
// ----------------- Designations -----------------
designations := r.Group("/api/designations")
designations.Use(middleware.AuthMiddleware(h))
{
    designations.POST("/", h.CreateDesignation)      // Create designation (ADMIN, SUPERADMIN, HR)
    designations.GET("/", h.GetAllDesignations)      // Get all designations (All authenticated users)
    designations.GET("/:id", h.GetDesignationByID)   // Get designation by ID (All authenticated users)
    designations.PATCH("/:id", h.UpdateDesignation)  // Update designation (ADMIN, SUPERADMIN, HR)
    designations.DELETE("/:id", h.DeleteDesignation) // Delete designation (ADMIN, SUPERADMIN, HR)
}
```

**Employee Route Added:**
```go
employees.PATCH("/:id/designation", h.UpdateEmployeeDesignation) // Assign/update designation (SUPERADMIN, ADMIN, HR)
```

---

## Employee Filtering

### Filter Employees by Role and Designation
**Endpoint:** `GET /api/employee`

**Query Parameters:**
- `role` (optional) - Filter by employee role (e.g., EMPLOYEE, MANAGER, HR, ADMIN, SUPERADMIN)
- `designation` (optional) - Filter by designation name (e.g., Senior Developer, Team Lead)

**Examples:**

1. **Get all employees:**
```bash
GET /api/employee
```

2. **Filter by role only:**
```bash
GET /api/employee?role=EMPLOYEE
```

3. **Filter by designation only:**
```bash
GET /api/employee?designation=Senior Developer
```

4. **Filter by both role and designation:**
```bash
GET /api/employee?role=EMPLOYEE&designation=Senior Developer
```

**Response:**
```json
{
  "message": "Employees fetched",
  "employees": [...],
  "filters": {
    "role": "EMPLOYEE",
    "designation": "Senior Developer"
  }
}
```

---

## Employee Response Structure

### Updated Employee Response
When fetching employees, the response now includes designation information:

```json
{
  "id": "employee-uuid",
  "full_name": "John Doe",
  "email": "john.doe@zenithive.com",
  "role": "EMPLOYEE",
  "status": "active",
  "manager_id": "manager-uuid",
  "manager_name": "Jane Smith",
  "designation_id": "designation-uuid",
  "designation_name": "Senior Developer",
  "salary": 75000,
  "joining_date": "2024-01-15T00:00:00Z",
  "ending_date": null,
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-12-04T08:00:00Z"
}
```

**Affected Endpoints:**
- `GET /api/employee` - All employees
- `GET /api/employee/:id` - Single employee
- `GET /api/employee/my-team` - Manager's team members

---

## Authorization Matrix

| Endpoint | EMPLOYEE | MANAGER | HR | ADMIN | SUPERADMIN |
|----------|----------|---------|-----|-------|------------|
| GET /api/designations | ✅ | ✅ | ✅ | ✅ | ✅ |
| GET /api/designations/:id | ✅ | ✅ | ✅ | ✅ | ✅ |
| POST /api/designations | ❌ | ❌ | ✅ | ✅ | ✅ |
| PATCH /api/designations/:id | ❌ | ❌ | ✅ | ✅ | ✅ |
| DELETE /api/designations/:id | ❌ | ❌ | ✅ | ✅ | ✅ |
| PATCH /api/employee/:id/designation | ❌ | ❌ | ✅ | ✅ | ✅ |

**Special Rules:**
- HR and ADMIN cannot modify SUPERADMIN users' designations
- Only SUPERADMIN can assign/modify designations for SUPERADMIN users

---

## Key Features

### 1. Cascade Behavior
- **ON DELETE SET NULL**: When a designation is deleted, employees keep their records but `designation_id` becomes `NULL`
- No data loss for employee records
- Safe deletion of designations

### 2. Data Integrity
- Foreign key constraint ensures designation exists before assignment
- Validation at controller level
- Database-level constraint enforcement

### 3. Role-Based Access Control
- View operations: All authenticated users
- Modify operations: ADMIN, SUPERADMIN, HR only
- SUPERADMIN protection: Only SUPERADMIN can modify SUPERADMIN users

### 4. Consistent Response Format
- Both ID and name returned for designations (like manager_id and manager_name)
- Useful for dropdowns (ID) and display (name)
- Optional fields (null if not assigned)

---

## Testing Checklist

### Designation CRUD
- [ ] Create designation as ADMIN
- [ ] Create designation as HR
- [ ] Create designation as SUPERADMIN
- [ ] Attempt to create as EMPLOYEE (should fail)
- [ ] Get all designations
- [ ] Get designation by ID
- [ ] Update designation
- [ ] Delete designation
- [ ] Verify employee designation_id becomes NULL after deletion

### Employee Designation Assignment
- [ ] Assign designation to employee
- [ ] Update employee designation
- [ ] Remove employee designation (set to null)
- [ ] Attempt to assign non-existent designation (should fail)
- [ ] HR attempts to assign designation to SUPERADMIN (should fail)
- [ ] SUPERADMIN assigns designation to SUPERADMIN (should succeed)

### Employee Queries
- [ ] Verify GET /api/employee returns designation_id and designation_name
- [ ] Verify GET /api/employee/:id returns designation_id and designation_name
- [ ] Verify GET /api/employee/my-team returns designation_id and designation_name
- [ ] Verify null designation_id shows as null (not error)

---

## Migration Instructions

### 1. Run Migration
```bash
goose -dir pkg/migration postgres "your-connection-string" up
```

### 2. Verify Tables
```sql
-- Check Tbl_Designation table
SELECT * FROM Tbl_Designation;

-- Check Tbl_Employee has designation_id column
\d Tbl_Employee;

-- Verify foreign key constraint
SELECT conname, conrelid::regclass, confrelid::regclass 
FROM pg_constraint 
WHERE conname = 'fk_employee_designation';
```

### 3. Test Cascade Behavior
```sql
-- Insert test designation
INSERT INTO Tbl_Designation (designation_name, description) 
VALUES ('Test Designation', 'Test description') 
RETURNING id;

-- Assign to employee
UPDATE Tbl_Employee 
SET designation_id = 'designation-uuid-here' 
WHERE id = 'employee-uuid-here';

-- Delete designation and verify employee designation_id becomes NULL
DELETE FROM Tbl_Designation WHERE id = 'designation-uuid-here';

-- Check employee record
SELECT id, full_name, designation_id FROM Tbl_Employee WHERE id = 'employee-uuid-here';
-- designation_id should be NULL
```

---

## Summary of Files Modified/Created

### New Files
1. `controllers/designation.go` - Designation controller with CRUD operations

### Modified Files
1. `pkg/migration/20251204073026_add_company_designation_responsibility.sql` - Database schema
2. `controllers/employee.go` - Added UpdateEmployeeDesignation function
3. `models/models.go` - Added Designation models and updated EmployeeInput
4. `repositories/repo.go` - Added designation repository functions and updated employee queries
5. `routes/router.go` - Added designation routes and employee designation route

---

## API Usage Examples

### Example 1: Create and Assign Designation
```bash
# 1. Create designation
curl -X POST http://localhost:8080/api/designations \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "designation_name": "Senior Developer",
    "description": "Senior software developer position"
  }'

# Response: {"message":"designation created successfully","designation_id":"uuid-here"}

# 2. Assign to employee
curl -X PATCH http://localhost:8080/api/employee/EMPLOYEE_ID/designation \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "designation_id": "uuid-here"
  }'

# 3. Get employee details
curl -X GET http://localhost:8080/api/employee/EMPLOYEE_ID \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Example 2: List All Designations
```bash
curl -X GET http://localhost:8080/api/designations \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Example 3: Update Designation
```bash
curl -X PATCH http://localhost:8080/api/designations/DESIGNATION_ID \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "designation_name": "Lead Developer",
    "description": "Updated description"
  }'
```

### Example 4: Remove Employee Designation
```bash
curl -X PATCH http://localhost:8080/api/employee/EMPLOYEE_ID/designation \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "designation_id": null
  }'
```

---

## Error Handling

### Common Error Responses

**401 Unauthorized:**
```json
{
  "error": "only ADMIN, SUPERADMIN, and HR can create designations"
}
```

**403 Forbidden:**
```json
{
  "error": "HR and ADMIN cannot modify SUPERADMIN users"
}
```

**404 Not Found:**
```json
{
  "error": "designation not found"
}
```

**400 Bad Request:**
```json
{
  "error": "invalid designation ID"
}
```

---

## Notes

1. **Backward Compatibility**: Existing employees without designations will have `designation_id` and `designation_name` as `null`
2. **Performance**: Designation name is fetched separately for each employee (consider JOIN optimization for large datasets)
3. **Validation**: Designation name is required, description is optional
4. **Security**: All endpoints require authentication via JWT token

---

## Future Enhancements (Optional)

1. Add department management and link to designations
2. Add grade levels for designations
3. Add designation history tracking
4. Bulk assign designations to multiple employees
5. Designation-based reporting and analytics
6. Designation approval workflow

---

**Document Version:** 1.0  
**Last Updated:** December 4, 2024  
**Author:** Development Team
