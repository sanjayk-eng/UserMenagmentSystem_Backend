# API Documentation - User Management System

## Base URL
```
http://localhost:{APP_PORT}/api
```

## Authentication
Most endpoints require JWT authentication. Include the token in the Authorization header:
```
Authorization: Bearer <your_jwt_token>
```

## Table of Contents
1. [Authentication](#authentication-endpoints)
2. [Employees](#employee-endpoints)
3. [Leaves](#leave-endpoints)
4. [Leave Balances](#leave-balance-endpoints)
5. [Payroll](#payroll-endpoints)
6. [Settings](#settings-endpoints)
7. [Holidays](#holiday-endpoints)
8. [Error Responses](#error-responses)

---

## Authentication Endpoints

### 1. Login
**POST** `/api/auth/login`

Authenticate user and receive JWT token.

**Request Body:**
```json
{
  "email": "user@zenithive.com",
  "password": "password123"
}
```

**Success Response (200):**
```json
{
  "success": true,
  "message": "Login successful",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@zenithive.com",
    "role": "EMPLOYEE"
  }
}
```

**Error Responses:**
- **400 Bad Request:** Invalid request payload
- **401 Unauthorized:** Login failed — email not found or wrong password

**cURL Example:**
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@zenithive.com",
    "password": "admin123"
  }'
```

---

## Employee Endpoints

### 1. Get All Employees
**GET** `/api/employee/`

Fetch all employees (SUPERADMIN, Admin only).

**Headers:**
```
Authorization: Bearer <token>
```

**Success Response (200):**
```json
{
  "message": "Employees fetched",
  "employees": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "full_name": "John Doe",
      "email": "john@zenithive.com",
      "status": "active",
      "role": "EMPLOYEE",
      "manager_id": "660e8400-e29b-41d4-a716-446655440001",
      "salary": 50000.00,
      "joining_date": "2024-01-15T00:00:00Z",
      "created_at": "2024-01-15T10:00:00Z",
      "updated_at": "2024-01-15T10:00:00Z",
      "deleted_at": null
    }
  ]
}
```

**Error Responses:**
- **401 Unauthorized:** not permitted
- **500 Internal Server Error:** Database error

**cURL Example:**
```bash
curl -X GET http://localhost:8080/api/employee/ \
  -H "Authorization: Bearer <your_token>"
```

### 2. Create Employee
**POST** `/api/employee/`

Create a new employee (SUPERADMIN, Admin only).

**Headers:**
```
Authorization: Bearer <token>
```

**Request Body:**
```json
{
  "full_name": "Jane Smith",
  "email": "jane@zenithive.com",
  "role": "EMPLOYEE",
  "password": "password123",
  "salary": 45000.00,
  "joining_date": "2024-11-25T00:00:00Z"
}
```

**Success Response (201):**
```json
{
  "message": "employee created"
}
```

**Error Responses:**
- **400 Bad Request:** Invalid input, email must end with @zenithive.com, email already exists, role not found
- **401 Unauthorized:** not permitted
- **500 Internal Server Error:** failed to create employee

**cURL Example:**
```bash
curl -X POST http://localhost:8080/api/employee/ \
  -H "Authorization: Bearer <your_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "full_name": "Jane Smith",
    "email": "jane@zenithive.com",
    "role": "EMPLOYEE",
    "password": "password123",
    "salary": 45000.00,
    "joining_date": "2024-11-25T00:00:00Z"
  }'
```

### 3. Update Employee Role
**PATCH** `/api/employee/:id/role`

Update an employee's role (SUPERADMIN, ADMIN only).

**Headers:**
```
Authorization: Bearer <token>
```

**URL Parameters:**
- `id` (UUID): Employee ID

**Request Body:**
```json
{
  "role": "MANAGER"
}
```

**Success Response (200):**
```json
{
  "message": "role updated",
  "employee_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Error Responses:**
- **401 Unauthorized:** not permitted
- **500 Internal Server Error:** Database error

**cURL Example:**
```bash
curl -X PATCH http://localhost:8080/api/employee/550e8400-e29b-41d4-a716-446655440000/role \
  -H "Authorization: Bearer <your_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "role": "MANAGER"
  }'
```

### 4. Update Employee Manager
**PATCH** `/api/employee/:id/manager`

Assign or change an employee's manager (SUPERADMIN, ADMIN, HR only).

**Headers:**
```
Authorization: Bearer <token>
```

**URL Parameters:**
- `id` (UUID): Employee ID

**Request Body:**
```json
{
  "manager_id": "660e8400-e29b-41d4-a716-446655440001"
}
```

**Success Response (200):**
```json
{
  "message": "manager updated",
  "employee_id": "550e8400-e29b-41d4-a716-446655440000",
  "manager_id": "660e8400-e29b-41d4-a716-446655440001"
}
```

**Error Responses:**
- **400 Bad Request:** cannot assign self
- **401 Unauthorized:** not permitted
- **404 Not Found:** manager not found
- **500 Internal Server Error:** failed

**cURL Example:**
```bash
curl -X PATCH http://localhost:8080/api/employee/550e8400-e29b-41d4-a716-446655440000/manager \
  -H "Authorization: Bearer <your_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "manager_id": "660e8400-e29b-41d4-a716-446655440001"
  }'
```

### 5. Get Employee Reports
**GET** `/api/employee/:id/reports`

Get direct reports of an employee (Self/Manager/Admin).

**Headers:**
```
Authorization: Bearer <token>
```

**URL Parameters:**
- `id` (UUID): Employee ID

**Success Response (200):**
```json
{
  "message": "Get employee reports"
}
```

**cURL Example:**
```bash
curl -X GET http://localhost:8080/api/employee/550e8400-e29b-41d4-a716-446655440000/reports \
  -H "Authorization: Bearer <your_token>"
```

---

## Leave Endpoints

### 1. Apply Leave
**POST** `/api/leaves/apply`

Employee applies for leave.

**Headers:**
```
Authorization: Bearer <token>
```

**Request Body:**
```json
{
  "leave_type_id": 1,
  "start_date": "2024-12-01T00:00:00Z",
  "end_date": "2024-12-05T00:00:00Z"
}
```

**Success Response (200):**
```json
{
  "message": "Leave applied successfully",
  "leave_id": "770e8400-e29b-41d4-a716-446655440002",
  "days": 5
}
```

**Error Responses:**
- **400 Bad Request:** Invalid input, Manager not assigned, Leave days must be greater than 0, Invalid leave type, Insufficient leave balance, Overlapping leave exists
- **401 Unauthorized:** Employee ID missing
- **403 Forbidden:** Only employees can apply leave
- **500 Internal Server Error:** Failed to start transaction, Failed to calculate leave days, Failed to fetch leave type, Failed to create leave balance, Failed to fetch leave balance, Failed to check overlapping leave, Failed to apply leave, Failed to commit transaction

**cURL Example:**
```bash
curl -X POST http://localhost:8080/api/leaves/apply \
  -H "Authorization: Bearer <your_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "leave_type_id": 1,
    "start_date": "2024-12-01T00:00:00Z",
    "end_date": "2024-12-05T00:00:00Z"
  }'
```

### 2. Admin Add Leave
**POST** `/api/leaves/admin-add`

Admin/Manager adds leave on behalf of employee.

**Headers:**
```
Authorization: Bearer <token>
```

**Request Body:**
```json
{
  "employee_id": "550e8400-e29b-41d4-a716-446655440000",
  "leave_type_id": 1,
  "start_date": "2024-12-10T00:00:00Z",
  "end_date": "2024-12-12T00:00:00Z"
}
```

**Success Response (200):**
```json
{
  "message": "Leave added successfully",
  "leave_id": "880e8400-e29b-41d4-a716-446655440003",
  "days": 3
}
```

**Error Responses:**
- **400 Bad Request:** Invalid input, Employee ID is required, Employee not found, Invalid leave type, Leave days must be greater than 0
- **401 Unauthorized:** not permitted to add leave
- **403 Forbidden:** Managers can only add leave for their team members
- **500 Internal Server Error:** failed to fetch company settings, Failed to start transaction, Failed to calculate leave days, Failed to fetch leave type, Failed to create leave balance, Failed to fetch leave balance, Failed to insert leave, Failed to update leave balance, Failed to commit transaction

**cURL Example:**
```bash
curl -X POST http://localhost:8080/api/leaves/admin-add \
  -H "Authorization: Bearer <your_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "employee_id": "550e8400-e29b-41d4-a716-446655440000",
    "leave_type_id": 1,
    "start_date": "2024-12-10T00:00:00Z",
    "end_date": "2024-12-12T00:00:00Z"
  }'
```

### 3. Add Leave Policy
**POST** `/api/leaves/admin-add/policy`

Create a new leave type/policy (SUPERADMIN only).

**Headers:**
```
Authorization: Bearer <token>
```

**Request Body:**
```json
{
  "name": "Sick Leave",
  "is_paid": true,
  "default_entitlement": 10,
  "leave_count": 2
}
```

**Success Response (200):**
```json
{
  "id": 3,
  "name": "Sick Leave",
  "is_paid": true,
  "default_entitlement": 10,
  "created_at": "2024-11-25T10:00:00Z",
  "updated_at": "2024-11-25T10:00:00Z"
}
```

**Error Responses:**
- **400 Bad Request:** Invalid input, leave_count must be greater than 0
- **401 Unauthorized:** not permitted to assign manager
- **500 Internal Server Error:** failed to get role, Failed to insert leave type

**cURL Example:**
```bash
curl -X POST http://localhost:8080/api/leaves/admin-add/policy \
  -H "Authorization: Bearer <your_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Sick Leave",
    "is_paid": true,
    "default_entitlement": 10,
    "leave_count": 2
  }'
```

### 4. Approve/Reject Leave
**POST** `/api/leaves/:id/action`

Approve or reject a leave request (Manager/Admin/SUPERADMIN only).

**Headers:**
```
Authorization: Bearer <token>
```

**URL Parameters:**
- `id` (UUID): Leave ID

**Request Body:**
```json
{
  "action": "APPROVE"
}
```
or
```json
{
  "action": "REJECT"
}
```

**Success Response (200):**
```json
{
  "message": "Leave approved successfully"
}
```
or
```json
{
  "message": "Leave rejected"
}
```

**Error Responses:**
- **400 Bad Request:** Invalid leave ID, Invalid payload, Action must be APPROVE or REJECT, Leave already processed
- **403 Forbidden:** Employees cannot approve leaves
- **404 Not Found:** Leave not found
- **500 Internal Server Error:** Failed to start transaction, Failed to reject leave, Failed to approve leave, Failed to update leave balance

**cURL Example:**
```bash
curl -X POST http://localhost:8080/api/leaves/770e8400-e29b-41d4-a716-446655440002/action \
  -H "Authorization: Bearer <your_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "action": "APPROVE"
  }'
```

### 5. Get All Leaves
**GET** `/api/leaves/all`

Get all leaves based on role:
- **EMPLOYEE**: Only their own leaves
- **MANAGER**: Leaves of their team members
- **ADMIN/SUPERADMIN**: All leaves

**Headers:**
```
Authorization: Bearer <token>
```

**Success Response (200):**
```json
{
  "total": 2,
  "data": [
    {
      "id": "770e8400-e29b-41d4-a716-446655440002",
      "employee": "John Doe",
      "leave_type": "Annual Leave",
      "start_date": "2024-12-01T00:00:00Z",
      "end_date": "2024-12-05T00:00:00Z",
      "days": 5,
      "status": "Pending",
      "applying_date": "2024-11-25T10:00:00Z"
    },
    {
      "id": "880e8400-e29b-41d4-a716-446655440003",
      "employee": "Jane Smith",
      "leave_type": "Sick Leave",
      "start_date": "2024-12-10T00:00:00Z",
      "end_date": "2024-12-12T00:00:00Z",
      "days": 3,
      "status": "APPROVED",
      "applying_date": "2024-11-24T15:30:00Z"
    }
  ]
}
```

**Error Responses:**
- **500 Internal Server Error:** Failed to fetch leaves

**cURL Example:**
```bash
curl -X GET http://localhost:8080/api/leaves/all \
  -H "Authorization: Bearer <your_token>"
```

---

## Leave Balance Endpoints

### 1. Get Leave Balances
**GET** `/api/leave-balances/employee/:id`

Get leave balances for an employee.

**Headers:**
```
Authorization: Bearer <token>
```

**URL Parameters:**
- `id` (UUID): Employee ID

**Success Response (200):**
```json
{
  "employee_id": "550e8400-e29b-41d4-a716-446655440000",
  "balances": [
    {
      "leave_type": "Annual Leave",
      "used": 5,
      "total": 20,
      "available": 15
    },
    {
      "leave_type": "Sick Leave",
      "used": 2,
      "total": 10,
      "available": 8
    }
  ]
}
```

**Error Responses:**
- **400 Bad Request:** Invalid employee ID
- **403 Forbidden:** Employees can only view their own balances
- **500 Internal Server Error:** Failed to fetch leave balances

**cURL Example:**
```bash
curl -X GET http://localhost:8080/api/leave-balances/employee/550e8400-e29b-41d4-a716-446655440000 \
  -H "Authorization: Bearer <your_token>"
```

### 2. Adjust Leave Balance
**POST** `/api/leave-balances/:id/adjust`

Manually adjust an employee's leave balance (ADMIN/SUPERADMIN only).

**Headers:**
```
Authorization: Bearer <token>
```

**URL Parameters:**
- `id` (UUID): Employee ID

**Request Body:**
```json
{
  "leave_type_id": 1,
  "quantity": 5,
  "reason": "Bonus leave for exceptional performance"
}
```
Note: Use negative quantity to deduct leaves.

**Success Response (200):**
```json
{
  "message": "Leave balance adjusted successfully",
  "new_adjusted": 5,
  "new_closing": 25,
  "year": 2024
}
```

**Error Responses:**
- **400 Bad Request:** Invalid employee ID, Invalid input
- **403 Forbidden:** Not authorized to adjust leave balances
- **500 Internal Server Error:** Failed to start transaction, Failed to fetch leave type, Failed to create leave balance, Failed to fetch leave balance, Failed to update leave balance, Failed to record leave adjustment, Transaction commit failed

**cURL Example:**
```bash
curl -X POST http://localhost:8080/api/leave-balances/550e8400-e29b-41d4-a716-446655440000/adjust \
  -H "Authorization: Bearer <your_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "leave_type_id": 1,
    "quantity": 5,
    "reason": "Bonus leave for exceptional performance"
  }'
```

---

## Payroll Endpoints

### 1. Run Payroll
**POST** `/api/payroll/run`

Generate payroll preview for a specific month and year (ADMIN/SUPERADMIN only).

**Headers:**
```
Authorization: Bearer <token>
```

**Request Body:**
```json
{
  "month": 11,
  "year": 2024
}
```

**Success Response (200):**
```json
{
  "payroll_run_id": "990e8400-e29b-41d4-a716-446655440004",
  "month": 11,
  "year": 2024,
  "total_payroll": 285000.00,
  "total_deductions": 15000.00,
  "employees_count": 6,
  "payroll_preview": [
    {
      "employee_id": "550e8400-e29b-41d4-a716-446655440000",
      "employee": "John Doe",
      "basic_salary": 50000.00,
      "working_days": 22,
      "absent_days": 2,
      "deductions": 4545.45,
      "net_salary": 45454.55
    }
  ]
}
```

**Error Responses:**
- **400 Bad Request:** Invalid input, Month must be between 1 and 12, Cannot run payroll for past months in the current year
- **403 Forbidden:** Not authorized to run payroll
- **500 Internal Server Error:** Role information missing, Invalid role type, Failed to fetch employees, Failed to calculate absent days, Failed to create payroll run

**cURL Example:**
```bash
curl -X POST http://localhost:8082/api/payroll/run \
  -H "Authorization: Bearer <your_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "month": 11,
    "year": 2024
  }'
```

### 2. Finalize Payroll
**POST** `/api/payroll/:id/finalize`

Finalize payroll and generate payslips (ADMIN/SUPERADMIN only).

**Headers:**
```
Authorization: Bearer <token>
```

**URL Parameters:**
- `id` (UUID): Payroll Run ID

**Success Response (200):**
```json
{
  "message": "Payroll finalized",
  "payslip_ids": [
    "aa0e8400-e29b-41d4-a716-446655440005",
    "bb0e8400-e29b-41d4-a716-446655440006"
  ],
  "working_days_used": 22
}
```

**Error Responses:**
- **400 Bad Request:** Invalid payroll run ID, Payroll already finalized
- **403 Forbidden:** Not authorized to finalize payroll
- **404 Not Found:** Payroll run not found
- **500 Internal Server Error:** Failed to start transaction, Failed to fetch employees, Failed to calculate absent days, Failed to insert payslip, Failed to update payroll run

**cURL Example:**
```bash
curl -X POST http://localhost:8080/api/payroll/990e8400-e29b-41d4-a716-446655440004/finalize \
  -H "Authorization: Bearer <your_token>"
```

### 3. Download Payslip PDF
**GET** `/api/payroll/payslips/:id/pdf`

Download payslip as PDF.

**Headers:**
```
Authorization: Bearer <token>
```

**URL Parameters:**
- `id` (UUID): Payslip ID

**Success Response (200):**
Returns PDF file for download.

**Error Responses:**
- **400 Bad Request:** Invalid payslip ID
- **404 Not Found:** Payslip not found
- **500 Internal Server Error:** Failed to generate PDF

**cURL Example:**
```bash
curl -X GET http://localhost:8080/api/payroll/payslips/aa0e8400-e29b-41d4-a716-446655440005/pdf \
  -H "Authorization: Bearer <your_token>" \
  --output payslip.pdf
```

---

## Settings Endpoints

### 1. Get Company Settings
**GET** `/api/settings/`

Get current company settings (ADMIN/SUPERADMIN only).

**Headers:**
```
Authorization: Bearer <token>
```

**Success Response (200):**
```json
{
  "settings": {
    "id": "cc0e8400-e29b-41d4-a716-446655440007",
    "working_days_per_month": 22,
    "allow_manager_add_leave": true,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-11-25T10:00:00Z"
  }
}
```

**Error Responses:**
- **403 Forbidden:** Not authorized to view settings
- **500 Internal Server Error:** Failed to fetch settings

**cURL Example:**
```bash
curl -X GET http://localhost:8080/api/settings/ \
  -H "Authorization: Bearer <your_token>"
```

### 2. Update Company Settings
**PUT** `/api/settings/`

Update company settings (ADMIN/SUPERADMIN only).

**Headers:**
```
Authorization: Bearer <token>
```

**Request Body:**
```json
{
  "working_days_per_month": 22,
  "allow_manager_add_leave": true
}
```

**Success Response (200):**
```json
{
  "message": "Company settings updated successfully"
}
```

**Error Responses:**
- **400 Bad Request:** Invalid input
- **403 Forbidden:** Not authorized to update settings
- **500 Internal Server Error:** Failed to update settings

**cURL Example:**
```bash
curl -X PUT http://localhost:8080/api/settings/ \
  -H "Authorization: Bearer <your_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "working_days_per_month": 22,
    "allow_manager_add_leave": true
  }'
```

---

## Holiday Endpoints

### 1. Add Holiday
**POST** `/api/settings/holidays/`

Add a new holiday (SUPERADMIN only).

**Headers:**
```
Authorization: Bearer <token>
```

**Request Body:**
```json
{
  "name": "Christmas",
  "date": "2024-12-25T00:00:00Z",
  "type": "HOLIDAY"
}
```

**Success Response (200):**
```json
{
  "message": "Holiday added successfully",
  "id": "dd0e8400-e29b-41d4-a716-446655440008"
}
```

**Error Responses:**
- **400 Bad Request:** Invalid input
- **401 Unauthorized:** not permitted
- **500 Internal Server Error:** Failed to add holiday

**cURL Example:**
```bash
curl -X POST http://localhost:8080/api/settings/holidays/ \
  -H "Authorization: Bearer <your_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Christmas",
    "date": "2024-12-25T00:00:00Z",
    "type": "HOLIDAY"
  }'
```

### 2. Get All Holidays
**GET** `/api/settings/holidays/`

Get all holidays (SUPERADMIN only).

**Headers:**
```
Authorization: Bearer <token>
```

**Success Response (200):**
```json
[
  {
    "id": "dd0e8400-e29b-41d4-a716-446655440008",
    "name": "Christmas",
    "date": "2024-12-25T00:00:00Z",
    "day": "Wednesday",
    "type": "HOLIDAY",
    "created_at": "2024-11-25T10:00:00Z",
    "updated_at": "2024-11-25T10:00:00Z"
  }
]
```

**Error Responses:**
- **401 Unauthorized:** not permitted
- **500 Internal Server Error:** Failed to fetch holidays

**cURL Example:**
```bash
curl -X GET http://localhost:8080/api/settings/holidays/ \
  -H "Authorization: Bearer <your_token>"
```

### 3. Delete Holiday
**DELETE** `/api/settings/holidays/:id`

Delete a holiday (SUPERADMIN only).

**Headers:**
```
Authorization: Bearer <token>
```

**URL Parameters:**
- `id` (UUID): Holiday ID

**Success Response (200):**
```json
{
  "message": "Holiday deleted successfully"
}
```

**Error Responses:**
- **400 Bad Request:** Holiday ID is required
- **401 Unauthorized:** not permitted
- **500 Internal Server Error:** Failed to delete holiday

**cURL Example:**
```bash
curl -X DELETE http://localhost:8080/api/settings/holidays/dd0e8400-e29b-41d4-a716-446655440008 \
  -H "Authorization: Bearer <your_token>"
```

---

## Error Responses

All error responses follow this format:

```json
{
  "error": {
    "code": 400,
    "message": "Detailed error message"
  }
}
```

### Common HTTP Status Codes

| Status Code | Description |
|-------------|-------------|
| 200 | OK - Request successful |
| 201 | Created - Resource created successfully |
| 400 | Bad Request - Invalid input or validation error |
| 401 | Unauthorized - Missing or invalid authentication token |
| 403 | Forbidden - User doesn't have permission |
| 404 | Not Found - Resource not found |
| 500 | Internal Server Error - Server-side error |

---

## Role-Based Access Control

### Roles
- **SUPERADMIN**: Full system access
- **ADMIN**: Administrative access (similar to SUPERADMIN but may have some restrictions)
- **HR**: Human resources operations
- **MANAGER**: Team management and leave approvals
- **EMPLOYEE**: Basic employee operations

### Permission Matrix

| Endpoint | SUPERADMIN | ADMIN | HR | MANAGER | EMPLOYEE |
|----------|------------|-------|-----|---------|----------|
| Login | ✅ | ✅ | ✅ | ✅ | ✅ |
| Get All Employees | ✅ | ✅ | ❌ | ❌ | ❌ |
| Create Employee | ✅ | ✅ | ❌ | ❌ | ❌ |
| Update Employee Role | ✅ | ✅ | ❌ | ❌ | ❌ |
| Update Employee Manager | ✅ | ✅ | ✅ | ❌ | ❌ |
| Apply Leave | ❌ | ❌ | ❌ | ❌ | ✅ |
| Admin Add Leave | ✅ | ❌ | ❌ | ✅* | ❌ |
| Add Leave Policy | ✅ | ❌ | ❌ | ❌ | ❌ |
| Approve/Reject Leave | ✅ | ✅ | ✅ | ✅ | ❌ |
| Get All Leaves | ✅ (all) | ✅ (all) | ✅ (all) | ✅ (team) | ✅ (own) |
| Get Leave Balances | ✅ (all) | ✅ (all) | ✅ (all) | ✅ (all) | ✅ (own) |
| Adjust Leave Balance | ✅ | ✅ | ❌ | ❌ | ❌ |
| Run Payroll | ✅ | ✅ | ❌ | ❌ | ❌ |
| Finalize Payroll | ✅ | ✅ | ❌ | ❌ | ❌ |
| Download Payslip | ✅ | ✅ | ✅ | ✅ | ✅ |
| Get/Update Settings | ✅ | ✅ | ❌ | ❌ | ❌ |
| Manage Holidays | ✅ | ❌ | ❌ | ❌ | ❌ |

*Manager can add leave only if `allow_manager_add_leave` setting is enabled

---

## Data Models

### Employee
```json
{
  "id": "UUID",
  "full_name": "string",
  "email": "string (must end with @zenithive.com)",
  "status": "string (active/inactive)",
  "role": "string (SUPERADMIN/ADMIN/HR/MANAGER/EMPLOYEE)",
  "manager_id": "UUID (nullable)",
  "salary": "float64",
  "joining_date": "timestamp",
  "created_at": "timestamp",
  "updated_at": "timestamp",
  "deleted_at": "timestamp (nullable)"
}
```

### Leave Type
```json
{
  "id": "integer",
  "name": "string",
  "is_paid": "boolean",
  "default_entitlement": "integer",
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

### Leave
```json
{
  "id": "UUID",
  "employee_id": "UUID",
  "leave_type_id": "integer",
  "start_date": "timestamp",
  "end_date": "timestamp",
  "days": "float64",
  "status": "string (Pending/APPROVED/REJECTED)",
  "applied_by": "UUID (nullable)",
  "approved_by": "UUID (nullable)",
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

### Leave Balance
```json
{
  "id": "UUID",
  "employee_id": "UUID",
  "leave_type_id": "integer",
  "year": "integer",
  "opening": "float64",
  "accrued": "float64",
  "used": "float64",
  "adjusted": "float64",
  "closing": "float64",
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

### Payroll Run
```json
{
  "id": "UUID",
  "month": "integer (1-12)",
  "year": "integer",
  "status": "string (PREVIEW/FINALIZED)",
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

### Payslip
```json
{
  "id": "UUID",
  "payroll_run_id": "UUID",
  "employee_id": "UUID",
  "basic_salary": "float64",
  "working_days": "integer",
  "absent_days": "integer",
  "deduction_amount": "float64",
  "net_salary": "float64",
  "pdf_path": "string (nullable)",
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

### Holiday
```json
{
  "id": "UUID",
  "name": "string",
  "date": "timestamp",
  "day": "string (auto-calculated)",
  "type": "string (default: HOLIDAY)",
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

### Company Settings
```json
{
  "id": "UUID",
  "working_days_per_month": "integer",
  "allow_manager_add_leave": "boolean",
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

---

## Business Logic Notes

### Leave Calculation
- **Working Days**: Only Monday-Friday are counted
- **Holidays**: Company holidays are excluded from leave calculations
- **Weekends**: Saturdays and Sundays are automatically excluded
- **Formula**: `working_days = count(Mon-Fri) - holidays between start_date and end_date`

### Payroll Calculation
- **Deduction Formula**: `(Basic Salary / Working Days) × Absent Days`
- **Net Salary**: `Basic Salary - Deductions`
- **Working Days**: Configurable in company settings (default: 22)
- **Absent Days**: Sum of approved leaves for the month

### Leave Balance
- **Opening**: Balance at the start of the year
- **Accrued**: Additional leaves earned during the year
- **Used**: Leaves taken (approved)
- **Adjusted**: Manual adjustments by admin
- **Closing**: `Opening + Accrued - Used + Adjusted`

### Leave Approval Workflow
1. Employee applies for leave (status: Pending)
2. Manager/Admin reviews the request
3. On approval: Leave status → APPROVED, balance updated
4. On rejection: Leave status → REJECTED, balance unchanged

---

## Environment Variables

Required environment variables in `.env`:

```env
APP_PORT=8080
FRONTEND_SERVER=http://localhost:3000
SERACT_KEY=your_jwt_secret_key
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=user_management_db
```

---

## Testing Examples

### Complete Workflow Example

#### 1. Login as Admin
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@zenithive.com",
    "password": "admin123"
  }'
```

Save the token from response.

#### 2. Create an Employee
```bash
curl -X POST http://localhost:8080/api/employee/ \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "full_name": "Test Employee",
    "email": "test@zenithive.com",
    "role": "EMPLOYEE",
    "password": "test123",
    "salary": 40000,
    "joining_date": "2024-11-01T00:00:00Z"
  }'
```

#### 3. Login as Employee
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@zenithive.com",
    "password": "test123"
  }'
```

#### 4. Apply for Leave
```bash
curl -X POST http://localhost:8080/api/leaves/apply \
  -H "Authorization: Bearer EMPLOYEE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "leave_type_id": 1,
    "start_date": "2024-12-01T00:00:00Z",
    "end_date": "2024-12-03T00:00:00Z"
  }'
```

#### 5. Approve Leave (as Manager/Admin)
```bash
curl -X POST http://localhost:8080/api/leaves/LEAVE_ID/action \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "action": "APPROVE"
  }'
```

#### 6. Run Payroll
```bash
curl -X POST http://localhost:8080/api/payroll/run \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "month": 12,
    "year": 2024
  }'
```

#### 7. Finalize Payroll
```bash
curl -X POST http://localhost:8080/api/payroll/PAYROLL_RUN_ID/finalize \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

#### 8. Download Payslip
```bash
curl -X GET http://localhost:8080/api/payroll/payslips/PAYSLIP_ID/pdf \
  -H "Authorization: Bearer EMPLOYEE_TOKEN" \
  --output payslip.pdf
```

---

## Notes

- All timestamps are in ISO 8601 format (UTC)
- UUIDs are in standard format: `550e8400-e29b-41d4-a716-446655440000`
- Email addresses must end with `@zenithive.com`
- JWT tokens expire based on server configuration
- All monetary values are in the base currency (no decimal places in storage, but displayed with 2 decimals)
- Leave days can be fractional (e.g., 2.5 days for half-day leaves)

---

**Last Updated**: November 25, 2024  
**API Version**: 1.0  
**Backend Framework**: Go (Gin)  
**Database**: PostgreSQL
