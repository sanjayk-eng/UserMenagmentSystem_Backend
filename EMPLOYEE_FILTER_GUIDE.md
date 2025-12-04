# Employee Filtering Guide

## Overview
The employee listing endpoint now supports filtering by role and designation name.

---

## Endpoint
**GET** `/api/employee`

**Access:** ADMIN, SUPERADMIN, HR only

---

## Query Parameters

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `role` | string | No | Filter by employee role | `EMPLOYEE`, `MANAGER`, `HR`, `ADMIN`, `SUPERADMIN` |
| `designation` | string | No | Filter by designation name | `Senior Developer`, `Team Lead` |

---

## Usage Examples

### 1. Get All Employees (No Filter)
```bash
GET /api/employee
```

**Response:**
```json
{
  "message": "Employees fetched",
  "employees": [
    {
      "id": "uuid-1",
      "full_name": "John Doe",
      "role": "EMPLOYEE",
      "designation_name": "Senior Developer"
    },
    {
      "id": "uuid-2",
      "full_name": "Jane Smith",
      "role": "MANAGER",
      "designation_name": "Team Lead"
    }
  ],
  "filters": {
    "role": "",
    "designation": ""
  }
}
```

---

### 2. Filter by Role Only
```bash
GET /api/employee?role=EMPLOYEE
```

**Response:**
```json
{
  "message": "Employees fetched",
  "employees": [
    {
      "id": "uuid-1",
      "full_name": "John Doe",
      "role": "EMPLOYEE",
      "designation_name": "Senior Developer"
    },
    {
      "id": "uuid-3",
      "full_name": "Bob Wilson",
      "role": "EMPLOYEE",
      "designation_name": "Junior Developer"
    }
  ],
  "filters": {
    "role": "EMPLOYEE",
    "designation": ""
  }
}
```

---

### 3. Filter by Designation Only
```bash
GET /api/employee?designation=Senior Developer
```

**Response:**
```json
{
  "message": "Employees fetched",
  "employees": [
    {
      "id": "uuid-1",
      "full_name": "John Doe",
      "role": "EMPLOYEE",
      "designation_name": "Senior Developer"
    },
    {
      "id": "uuid-4",
      "full_name": "Alice Brown",
      "role": "EMPLOYEE",
      "designation_name": "Senior Developer"
    }
  ],
  "filters": {
    "role": "",
    "designation": "Senior Developer"
  }
}
```

---

### 4. Filter by Both Role and Designation
```bash
GET /api/employee?role=EMPLOYEE&designation=Senior Developer
```

**Response:**
```json
{
  "message": "Employees fetched",
  "employees": [
    {
      "id": "uuid-1",
      "full_name": "John Doe",
      "role": "EMPLOYEE",
      "designation_name": "Senior Developer"
    }
  ],
  "filters": {
    "role": "EMPLOYEE",
    "designation": "Senior Developer"
  }
}
```

---

## cURL Examples

### Get all employees
```bash
curl -X GET "http://localhost:8080/api/employee" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Filter by role
```bash
curl -X GET "http://localhost:8080/api/employee?role=EMPLOYEE" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Filter by designation
```bash
curl -X GET "http://localhost:8080/api/employee?designation=Senior%20Developer" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Filter by both
```bash
curl -X GET "http://localhost:8080/api/employee?role=EMPLOYEE&designation=Senior%20Developer" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

## Valid Role Values

- `EMPLOYEE`
- `MANAGER`
- `HR`
- `ADMIN`
- `SUPERADMIN`

---

## Designation Values

Designation values are dynamic and based on what's created in the system. Use the designation name exactly as it appears in the database.

To get all available designations:
```bash
GET /api/designations
```

---

## Implementation Details

### Controller
**File:** `controllers/employee.go`

```go
func (h *HandlerFunc) GetEmployee(c *gin.Context) {
    // Get filter parameters from query string
    roleFilter := c.Query("role")
    designationFilter := c.Query("designation")
    
    employees, err := h.Query.GetAllEmployees(roleFilter, designationFilter)
    // ...
}
```

### Repository
**File:** `repositories/repo.go`

```go
func (r *Repository) GetAllEmployees(roleFilter, designationFilter string) ([]models.EmployeeInput, error) {
    // Dynamic query building with LEFT JOIN on Tbl_Designation
    // Filters applied using WHERE clauses
    // ...
}
```

---

## SQL Query Structure

The repository builds a dynamic SQL query:

```sql
SELECT 
    e.id, e.full_name, e.email, e.status,
    r.type AS role, e.password, e.manager_id, e.designation_id,
    e.salary, e.joining_date, e.ending_date,
    e.created_at, e.updated_at, e.deleted_at
FROM Tbl_Employee e
JOIN Tbl_Role r ON e.role_id = r.id
LEFT JOIN Tbl_Designation d ON e.designation_id = d.id
WHERE 1=1
    AND r.type = $1              -- if role filter provided
    AND d.designation_name = $2  -- if designation filter provided
ORDER BY e.full_name
```

**Key Points:**
- Uses `LEFT JOIN` to include employees without designations
- Filters are optional and applied dynamically
- Results are ordered by employee name

---

## Edge Cases

### 1. Employees Without Designation
When filtering by designation, employees without a designation are excluded:
```bash
GET /api/employee?designation=Senior Developer
# Only returns employees with "Senior Developer" designation
```

### 2. Invalid Role
If an invalid role is provided, the query returns an empty array:
```bash
GET /api/employee?role=INVALID_ROLE
# Returns: {"employees": [], "filters": {"role": "INVALID_ROLE"}}
```

### 3. Invalid Designation
If a designation doesn't exist, the query returns an empty array:
```bash
GET /api/employee?designation=Non Existent
# Returns: {"employees": [], "filters": {"designation": "Non Existent"}}
```

### 4. Case Sensitivity
Filters are case-sensitive. Ensure exact matches:
- ✅ `role=EMPLOYEE`
- ❌ `role=employee`
- ✅ `designation=Senior Developer`
- ❌ `designation=senior developer`

---

## Testing Checklist

- [ ] Get all employees without filters
- [ ] Filter by role only (each role type)
- [ ] Filter by designation only
- [ ] Filter by both role and designation
- [ ] Test with invalid role
- [ ] Test with invalid designation
- [ ] Test with employees who have no designation
- [ ] Verify case sensitivity
- [ ] Test URL encoding for designation names with spaces
- [ ] Verify authorization (only ADMIN, SUPERADMIN, HR can access)

---

## Frontend Integration Example

### JavaScript/Fetch
```javascript
// Get all employees
fetch('/api/employee', {
  headers: { 'Authorization': `Bearer ${token}` }
})

// Filter by role
fetch('/api/employee?role=EMPLOYEE', {
  headers: { 'Authorization': `Bearer ${token}` }
})

// Filter by designation
fetch('/api/employee?designation=Senior Developer', {
  headers: { 'Authorization': `Bearer ${token}` }
})

// Filter by both
fetch('/api/employee?role=EMPLOYEE&designation=Senior Developer', {
  headers: { 'Authorization': `Bearer ${token}` }
})
```

### React Example
```jsx
const [employees, setEmployees] = useState([]);
const [roleFilter, setRoleFilter] = useState('');
const [designationFilter, setDesignationFilter] = useState('');

const fetchEmployees = async () => {
  const params = new URLSearchParams();
  if (roleFilter) params.append('role', roleFilter);
  if (designationFilter) params.append('designation', designationFilter);
  
  const response = await fetch(`/api/employee?${params}`, {
    headers: { 'Authorization': `Bearer ${token}` }
  });
  const data = await response.json();
  setEmployees(data.employees);
};

// Dropdown for role filter
<select onChange={(e) => setRoleFilter(e.target.value)}>
  <option value="">All Roles</option>
  <option value="EMPLOYEE">Employee</option>
  <option value="MANAGER">Manager</option>
  <option value="HR">HR</option>
  <option value="ADMIN">Admin</option>
  <option value="SUPERADMIN">Super Admin</option>
</select>

// Dropdown for designation filter
<select onChange={(e) => setDesignationFilter(e.target.value)}>
  <option value="">All Designations</option>
  {designations.map(d => (
    <option key={d.id} value={d.designation_name}>
      {d.designation_name}
    </option>
  ))}
</select>
```

---

## Performance Considerations

1. **Indexing:** Consider adding indexes on frequently filtered columns:
   ```sql
   CREATE INDEX idx_employee_role ON Tbl_Employee(role_id);
   CREATE INDEX idx_employee_designation ON Tbl_Employee(designation_id);
   CREATE INDEX idx_designation_name ON Tbl_Designation(designation_name);
   ```

2. **Query Optimization:** The LEFT JOIN ensures employees without designations are included in unfiltered results

3. **Pagination:** For large datasets, consider adding pagination:
   ```bash
   GET /api/employee?role=EMPLOYEE&page=1&limit=50
   ```

---

## Error Responses

### 401 Unauthorized
```json
{
  "error": "not permitted"
}
```
**Cause:** User is not ADMIN, SUPERADMIN, or HR

### 500 Internal Server Error
```json
{
  "error": "database connection error"
}
```
**Cause:** Database query failed

---

## Summary

✅ **Added Features:**
- Filter employees by role
- Filter employees by designation name
- Combine both filters
- Response includes applied filters

✅ **Benefits:**
- Easier employee management
- Quick filtering for reports
- Better user experience
- Flexible querying

---

**Document Version:** 1.0  
**Last Updated:** December 4, 2024
