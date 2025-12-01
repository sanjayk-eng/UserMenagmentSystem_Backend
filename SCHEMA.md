# Database Schema & System Diagrams

## Table of Contents
1. [Entity Relationship Diagram](#entity-relationship-diagram)
2. [Database Tables](#database-tables)
3. [Activity Diagrams](#activity-diagrams)
4. [Sequence Diagrams](#sequence-diagrams)

---

## Entity Relationship Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         USER MANAGEMENT SYSTEM                               │
│                         DATABASE SCHEMA (PostgreSQL)                         │
└─────────────────────────────────────────────────────────────────────────────┘

┌──────────────────────┐
│     Tbl_Role         │
├──────────────────────┤
│ PK  id (SERIAL)      │
│     type (TEXT)      │◄────────────┐
│     created_at       │             │
│     updated_at       │             │
└──────────────────────┘             │
                                     │
                                     │ role_id (FK)
                                     │
┌──────────────────────────────────────────────┐
│           Tbl_Employee                       │
├──────────────────────────────────────────────┤
│ PK  id (UUID)                                │◄─────────┐
│ FK  role_id → Tbl_Role.id                    │          │
│ FK  manager_id → Tbl_Employee.id (self-ref)  │──────────┘
│     full_name (TEXT)                         │
│     email (TEXT) UNIQUE                      │
│     password (TEXT)                          │
│     salary (NUMERIC)                         │
│     joining_date (DATE)                      │
│     status (VARCHAR)                         │
│     deleted_at (TIMESTAMP)                   │
│     created_at (TIMESTAMP)                   │
│     updated_at (TIMESTAMP)                   │
└──────────────────────────────────────────────┘
         │                    │
         │                    │
         │                    └──────────────────────────┐
         │                                               │
         │ employee_id (FK)                              │ actor_id (FK)
         │                                               │
         ▼                                               ▼
┌──────────────────────┐                    ┌──────────────────────┐
│   Tbl_Leave          │                    │    Tbl_Audit         │
├──────────────────────┤                    ├──────────────────────┤
│ PK  id (UUID)        │                    │ PK  id (UUID)        │
│ FK  employee_id      │                    │ FK  actor_id         │
│ FK  leave_type_id    │                    │     action (TEXT)    │
│ FK  applied_by       │                    │     entity (TEXT)    │
│ FK  approved_by      │                    │     entity_id (UUID) │
│     start_date       │                    │     metadata (JSONB) │
│     end_date         │                    │     created_at       │
│     days (NUMERIC)   │                    │     updated_at       │
│     status (TEXT)    │                    └──────────────────────┘
│     created_at       │
│     updated_at       │
└──────────────────────┘
         │
         │ leave_type_id (FK)
         │
         ▼
┌──────────────────────┐
│  Tbl_Leave_type      │
├──────────────────────┤
│ PK  id (SERIAL)      │◄────────────┐
│     name (TEXT)      │             │
│     is_paid (BOOL)   │             │
│     default_entitle  │             │
│     created_at       │             │
│     updated_at       │             │
└──────────────────────┘             │
         │                           │
         │                           │ leave_type_id (FK)
         │                           │
         └───────────────────────────┼──────────────────┐
                                     │                  │
                                     │                  │
                                     ▼                  ▼
┌──────────────────────────────────────┐   ┌──────────────────────────────┐
│      Tbl_Leave_balance               │   │   Tbl_Leave_adjustment       │
├──────────────────────────────────────┤   ├──────────────────────────────┤
│ PK  id (UUID)                        │   │ PK  id (UUID)                │
│ FK  employee_id → Tbl_Employee.id    │   │ FK  employee_id              │
│ FK  leave_type_id → Tbl_Leave_type   │   │ FK  leave_type_id            │
│     year (INT)                       │   │ FK  created_by               │
│     opening (NUMERIC)                │   │     quantity (NUMERIC)       │
│     accrued (NUMERIC)                │   │     reason (TEXT)            │
│     used (NUMERIC)                   │   │     year (INT)               │
│     adjusted (NUMERIC)               │   │     created_at               │
│     closing (NUMERIC)                │   │     updated_at               │
│     created_at                       │   └──────────────────────────────┘
│     updated_at                       │
└──────────────────────────────────────┘


┌──────────────────────────────────────┐
│      Tbl_Payroll_run                 │
├──────────────────────────────────────┤
│ PK  id (UUID)                        │
│     month (INT)                      │
│     year (INT)                       │
│     status (TEXT)                    │
│     created_at                       │
│     updated_at                       │
└──────────────────────────────────────┘
         │
         │ payroll_run_id (FK)
         │
         ▼
┌──────────────────────────────────────┐
│         Tbl_Payslip                  │
├──────────────────────────────────────┤
│ PK  id (UUID)                        │
│ FK  payroll_run_id                   │
│ FK  employee_id → Tbl_Employee.id    │
│     basic_salary (NUMERIC)           │
│     working_days (INT)               │
│     absent_days (INT)                │
│     deduction_amount (NUMERIC)       │
│     net_salary (NUMERIC)             │
│     pdf_path (TEXT)                  │
│     created_at                       │
│     updated_at                       │
└──────────────────────────────────────┘


┌──────────────────────────────────────┐
│         Tbl_Holiday                  │
├──────────────────────────────────────┤
│ PK  id (SERIAL)                      │
│     name (TEXT)                      │
│     date (DATE) UNIQUE               │
│     day (TEXT)                       │
│     type (TEXT)                      │
│     created_at                       │
│     updated_at                       │
└──────────────────────────────────────┘


┌──────────────────────────────────────┐
│    Tbl_Company_Settings              │
├──────────────────────────────────────┤
│ PK  id (UUID)                        │
│     working_days_per_month (INT)     │
│     allow_manager_add_leave (BOOL)   │
│     created_at                       │
│     updated_at                       │
└──────────────────────────────────────┘
```

---

## Database Tables

### 1. Tbl_Role
Stores user roles in the system.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Unique role identifier |
| type | TEXT | NOT NULL, UNIQUE | Role name (SUPERADMIN, ADMIN, HR, MANAGER, EMPLOYEE) |
| created_at | TIMESTAMP | DEFAULT NOW() | Record creation time |
| updated_at | TIMESTAMP | DEFAULT NOW() | Last update time |

**Default Roles**:
- SUPERADMIN
- ADMIN
- HR
- MANAGER
- EMPLOYEE

---

### 2. Tbl_Employee
Stores employee information.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | Unique employee identifier |
| full_name | TEXT | NOT NULL | Employee full name |
| email | TEXT | UNIQUE, NOT NULL | Employee email (@zenithive.com) |
| role_id | INT | FK → Tbl_Role.id | Employee role |
| password | TEXT | NOT NULL | Hashed password |
| manager_id | UUID | FK → Tbl_Employee.id | Manager reference (self-referencing) |
| salary | NUMERIC | | Employee salary |
| joining_date | DATE | | Date of joining |
| ending_date | DATE | NULL | Date of leaving/resignation (optional) |
| status | VARCHAR(20) | DEFAULT 'active' | Employee status (active/inactive) |
| deleted_at | TIMESTAMP | NULL | Soft delete timestamp |
| created_at | TIMESTAMP | DEFAULT NOW() | Record creation time |
| updated_at | TIMESTAMP | DEFAULT NOW() | Last update time |

**Relationships**:
- Self-referencing: `manager_id` → `Tbl_Employee.id`
- Many-to-One: `role_id` → `Tbl_Role.id`

---

### 3. Tbl_Leave_type
Defines types of leaves available.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Unique leave type identifier |
| name | TEXT | NOT NULL | Leave type name (Annual, Sick, etc.) |
| is_paid | BOOLEAN | | Whether leave is paid |
| default_entitlement | INT | | Default number of days per year |
| created_at | TIMESTAMP | DEFAULT NOW() | Record creation time |
| updated_at | TIMESTAMP | DEFAULT NOW() | Last update time |

---

### 4. Tbl_Leave
Stores leave applications.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | Unique leave identifier |
| employee_id | UUID | FK → Tbl_Employee.id | Employee applying for leave |
| leave_type_id | INT | FK → Tbl_Leave_type.id | Type of leave |
| start_date | DATE | | Leave start date |
| end_date | DATE | | Leave end date |
| days | NUMERIC | | Number of leave days |
| status | TEXT | | Leave status (Pending/MANAGER_APPROVED/MANAGER_REJECTED/APPROVED/REJECTED/CANCELLED/WITHDRAWAL_PENDING/WITHDRAWN) |
| applied_by | UUID | FK → Tbl_Employee.id | Who applied (manager for admin-add) |
| approved_by | UUID | FK → Tbl_Employee.id | Who approved/rejected |
| created_at | TIMESTAMP | DEFAULT NOW() | Application time |
| updated_at | TIMESTAMP | DEFAULT NOW() | Last update time |

---

### 5. Tbl_Leave_balance
Tracks employee leave balances per year.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | Unique balance identifier |
| employee_id | UUID | FK → Tbl_Employee.id | Employee reference |
| leave_type_id | INT | FK → Tbl_Leave_type.id | Leave type reference |
| year | INT | | Year for this balance |
| opening | NUMERIC | | Opening balance |
| accrued | NUMERIC | | Leaves accrued during year |
| used | NUMERIC | | Leaves used |
| adjusted | NUMERIC | | Manual adjustments |
| closing | NUMERIC | | Closing balance (opening + accrued - used + adjusted) |
| created_at | TIMESTAMP | DEFAULT NOW() | Record creation time |
| updated_at | TIMESTAMP | DEFAULT NOW() | Last update time |

---

### 6. Tbl_Leave_adjustment
Logs manual leave balance adjustments.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | Unique adjustment identifier |
| employee_id | UUID | FK → Tbl_Employee.id | Employee reference |
| leave_type_id | INT | FK → Tbl_Leave_type.id | Leave type reference |
| quantity | NUMERIC | | Adjustment amount (+ve or -ve) |
| reason | TEXT | | Reason for adjustment |
| year | INT | | Year of adjustment |
| created_by | UUID | FK → Tbl_Employee.id | Admin who made adjustment |
| created_at | TIMESTAMP | DEFAULT NOW() | Adjustment time |
| updated_at | TIMESTAMP | DEFAULT NOW() | Last update time |

---

### 7. Tbl_Payroll_run
Stores payroll run information.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | Unique payroll run identifier |
| month | INT | | Month (1-12) |
| year | INT | | Year |
| status | TEXT | | Status (PREVIEW/FINALIZED) |
| created_at | TIMESTAMP | DEFAULT NOW() | Run creation time |
| updated_at | TIMESTAMP | DEFAULT NOW() | Last update time |

---

### 8. Tbl_Payslip
Stores individual employee payslips.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | Unique payslip identifier |
| payroll_run_id | UUID | FK → Tbl_Payroll_run.id | Payroll run reference |
| employee_id | UUID | FK → Tbl_Employee.id | Employee reference |
| basic_salary | NUMERIC | | Employee's basic salary |
| working_days | INT | | Working days in month |
| absent_days | INT | | Absent days (approved leaves) |
| deduction_amount | NUMERIC | | Deduction for absences |
| net_salary | NUMERIC | | Net salary after deductions |
| pdf_path | TEXT | | Path to generated PDF |
| created_at | TIMESTAMP | DEFAULT NOW() | Payslip creation time |
| updated_at | TIMESTAMP | DEFAULT NOW() | Last update time |

---

### 9. Tbl_Holiday
Stores company holidays.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Unique holiday identifier |
| name | TEXT | NOT NULL | Holiday name |
| date | DATE | UNIQUE, NOT NULL | Holiday date |
| day | TEXT | NOT NULL | Day of week (auto-calculated) |
| type | TEXT | NOT NULL | Holiday type (default: HOLIDAY) |
| created_at | TIMESTAMP | DEFAULT NOW() | Record creation time |
| updated_at | TIMESTAMP | DEFAULT NOW() | Last update time |

---

### 10. Tbl_Company_Settings
Stores company-wide settings.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | Unique settings identifier |
| working_days_per_month | INT | DEFAULT 22 | Working days per month |
| allow_manager_add_leave | BOOLEAN | DEFAULT FALSE | Allow managers to add leave |
| created_at | TIMESTAMP | DEFAULT NOW() | Record creation time |
| updated_at | TIMESTAMP | DEFAULT NOW() | Last update time |

---

### 11. Tbl_Audit
Stores audit logs for system actions.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | Unique audit identifier |
| actor_id | UUID | FK → Tbl_Employee.id | User who performed action |
| action | TEXT | | Action performed |
| entity | TEXT | | Entity affected |
| entity_id | UUID | | ID of affected entity |
| metadata | JSONB | | Additional metadata |
| created_at | TIMESTAMP | DEFAULT NOW() | Action time |
| updated_at | TIMESTAMP | DEFAULT NOW() | Last update time |

---

## Activity Diagrams

### 1. Employee Creation Activity Diagram

```
                    START
                      │
                      ▼
        ┌─────────────────────────┐
        │  Admin/HR Initiates     │
        │  Employee Creation      │
        └────────────┬────────────┘
                     │
                     ▼
        ┌─────────────────────────┐
        │  POST /api/employee/    │
        │  with employee data     │
        └────────────┬────────────┘
                     │
                     ▼
        ┌─────────────────────────┐
        │  Validate Input         │
        │  - Email format         │
        │  - Required fields      │
        │  - Email domain         │
        └────────────┬────────────┘
                     │
                     ▼
              ┌──────────────┐
              │  Valid?      │
              └──────┬───────┘
                     │
         ┌───────────┴───────────┐
         │ NO                    │ YES
         ▼                       ▼
┌─────────────────┐    ┌─────────────────────┐
│  Return 400     │    │  Check Email Exists │
│  Error          │    └──────────┬──────────┘
└─────────────────┘               │
                                  ▼
                         ┌─────────────────┐
                         │  Email Exists?  │
                         └────────┬────────┘
                                  │
                      ┌───────────┴───────────┐
                      │ YES                   │ NO
                      ▼                       ▼
            ┌─────────────────┐    ┌─────────────────────┐
            │  Return 400     │    │  Get Role ID        │
            │  Email exists   │    └──────────┬──────────┘
            └─────────────────┘               │
                                              ▼
                                   ┌─────────────────────┐
                                   │  Hash Password      │
                                   └──────────┬──────────┘
                                              │
                                              ▼
                                   ┌─────────────────────┐
                                   │  Insert Employee    │
                                   │  into Database      │
                                   └──────────┬──────────┘
                                              │
                                              ▼
                                   ┌─────────────────────┐
                                   │  Success?           │
                                   └──────────┬──────────┘
                                              │
                                  ┌───────────┴───────────┐
                                  │ NO                    │ YES
                                  ▼                       ▼
                        ┌─────────────────┐    ┌─────────────────────┐
                        │  Return 500     │    │  Spawn Goroutine    │
                        │  Error          │    │  (Async Email)      │
                        └─────────────────┘    └──────────┬──────────┘
                                                          │
                                                          ▼
                                               ┌─────────────────────┐
                                               │  Send Welcome Email │
                                               │  with Credentials   │
                                               └──────────┬──────────┘
                                                          │
                                                          ▼
                                               ┌─────────────────────┐
                                               │  Return 201         │
                                               │  Employee Created   │
                                               └──────────┬──────────┘
                                                          │
                                                          ▼
                                                        END
```

---

### 2. Leave Application Activity Diagram

```
                              START
                                │
                                ▼
                  ┌──────────────────────────┐
                  │  Employee Initiates      │
                  │  Leave Application       │
                  └────────────┬─────────────┘
                               │
                               ▼
                  ┌──────────────────────────┐
                  │  POST /api/leaves/apply  │
                  │  with leave details      │
                  └────────────┬─────────────┘
                               │
                               ▼
                  ┌──────────────────────────┐
                  │  Authenticate User       │
                  │  (JWT Middleware)        │
                  └────────────┬─────────────┘
                               │
                               ▼
                        ┌──────────────┐
                        │  Valid Token?│
                        └──────┬───────┘
                               │
                   ┌───────────┴───────────┐
                   │ NO                    │ YES
                   ▼                       ▼
          ┌─────────────────┐    ┌─────────────────────┐
          │  Return 401     │    │  Check Role         │
          │  Unauthorized   │    │  = EMPLOYEE?        │
          └─────────────────┘    └──────────┬──────────┘
                                            │
                                ┌───────────┴───────────┐
                                │ NO                    │ YES
                                ▼                       ▼
                      ┌─────────────────┐    ┌─────────────────────┐
                      │  Return 403     │    │  Validate Input     │
                      │  Forbidden      │    │  - Dates            │
                      └─────────────────┘    │  - Leave Type       │
                                             └──────────┬──────────┘
                                                        │
                                                        ▼
                                             ┌─────────────────────┐
                                             │  Start Transaction  │
                                             └──────────┬──────────┘
                                                        │
                                                        ▼
                                             ┌─────────────────────┐
                                             │  Fetch Manager ID   │
                                             └──────────┬──────────┘
                                                        │
                                                        ▼
                                                ┌───────────────┐
                                                │  Has Manager? │
                                                └───────┬───────┘
                                                        │
                                            ┌───────────┴───────────┐
                                            │ NO                    │ YES
                                            ▼                       ▼
                                  ┌─────────────────┐    ┌─────────────────────┐
                                  │  Rollback       │    │  Calculate Working  │
                                  │  Return 400     │    │  Days (skip         │
                                  └─────────────────┘    │  weekends/holidays) │
                                                         └──────────┬──────────┘
                                                                    │
                                                                    ▼
                                                         ┌─────────────────────┐
                                                         │  Days > 0?          │
                                                         └──────────┬──────────┘
                                                                    │
                                                        ┌───────────┴───────────┐
                                                        │ NO                    │ YES
                                                        ▼                       ▼
                                              ┌─────────────────┐    ┌─────────────────────┐
                                              │  Rollback       │    │  Validate Leave     │
                                              │  Return 400     │    │  Type Exists        │
                                              └─────────────────┘    └──────────┬──────────┘
                                                                                │
                                                                                ▼
                                                                     ┌─────────────────────┐
                                                                     │  Get/Create Leave   │
                                                                     │  Balance            │
                                                                     └──────────┬──────────┘
                                                                                │
                                                                                ▼
                                                                     ┌─────────────────────┐
                                                                     │  Sufficient         │
                                                                     │  Balance?           │
                                                                     └──────────┬──────────┘
                                                                                │
                                                                    ┌───────────┴───────────┐
                                                                    │ NO                    │ YES
                                                                    ▼                       ▼
                                                          ┌─────────────────┐    ┌─────────────────────┐
                                                          │  Rollback       │    │  Check Overlapping  │
                                                          │  Return 400     │    │  Leaves             │
                                                          └─────────────────┘    └──────────┬──────────┘
                                                                                            │
                                                                                            ▼
                                                                                 ┌─────────────────────┐
                                                                                 │  Overlap Exists?    │
                                                                                 └──────────┬──────────┘
                                                                                            │
                                                                                ┌───────────┴───────────┐
                                                                                │ YES                   │ NO
                                                                                ▼                       ▼
                                                                      ┌─────────────────┐    ┌─────────────────────┐
                                                                      │  Rollback       │    │  Insert Leave       │
                                                                      │  Return 400     │    │  (Status: Pending)  │
                                                                      └─────────────────┘    └──────────┬──────────┘
                                                                                                        │
                                                                                                        ▼
                                                                                             ┌─────────────────────┐
                                                                                             │  Commit Transaction │
                                                                                             └──────────┬──────────┘
                                                                                                        │
                                                                                                        ▼
                                                                                             ┌─────────────────────┐
                                                                                             │  Spawn Goroutine    │
                                                                                             │  (Async Email)      │
                                                                                             └──────────┬──────────┘
                                                                                                        │
                                                                                                        ▼
                                                                                             ┌─────────────────────┐
                                                                                             │  Fetch Manager &    │
                                                                                             │  Admin Emails       │
                                                                                             └──────────┬──────────┘
                                                                                                        │
                                                                                                        ▼
                                                                                             ┌─────────────────────┐
                                                                                             │  Send Notification  │
                                                                                             │  Emails             │
                                                                                             └──────────┬──────────┘
                                                                                                        │
                                                                                                        ▼
                                                                                             ┌─────────────────────┐
                                                                                             │  Return 200         │
                                                                                             │  Leave Applied      │
                                                                                             └──────────┬──────────┘
                                                                                                        │
                                                                                                        ▼
                                                                                                       END
```

---

### 3. Leave Approval/Rejection Activity Diagram

```
                                    START
                                      │
                                      ▼
                        ┌──────────────────────────┐
                        │  Manager/Admin Initiates │
                        │  Leave Action            │
                        └────────────┬─────────────┘
                                     │
                                     ▼
                        ┌──────────────────────────┐
                        │  POST /api/leaves/:id/   │
                        │  action                  │
                        │  {action: APPROVE/REJECT}│
                        └────────────┬─────────────┘
                                     │
                                     ▼
                        ┌──────────────────────────┐
                        │  Authenticate User       │
                        │  (JWT Middleware)        │
                        └────────────┬─────────────┘
                                     │
                                     ▼
                              ┌──────────────┐
                              │  Valid Token?│
                              └──────┬───────┘
                                     │
                         ┌───────────┴───────────┐
                         │ NO                    │ YES
                         ▼                       ▼
                ┌─────────────────┐    ┌─────────────────────┐
                │  Return 401     │    │  Check Role         │
                │  Unauthorized   │    │  != EMPLOYEE?       │
                └─────────────────┘    └──────────┬──────────┘
                                                  │
                                      ┌───────────┴───────────┐
                                      │ NO (is EMPLOYEE)      │ YES
                                      ▼                       ▼
                            ┌─────────────────┐    ┌─────────────────────┐
                            │  Return 403     │    │  Validate Action    │
                            │  Forbidden      │    │  (APPROVE/REJECT)   │
                            └─────────────────┘    └──────────┬──────────┘
                                                              │
                                                              ▼
                                                   ┌─────────────────────┐
                                                   │  Start Transaction  │
                                                   └──────────┬──────────┘
                                                              │
                                                              ▼
                                                   ┌─────────────────────┐
                                                   │  Fetch Leave Record │
                                                   │  (with lock)        │
                                                   └──────────┬──────────┘
                                                              │
                                                              ▼
                                                       ┌──────────────┐
                                                       │  Found?      │
                                                       └──────┬───────┘
                                                              │
                                                  ┌───────────┴───────────┐
                                                  │ NO                    │ YES
                                                  ▼                       ▼
                                        ┌─────────────────┐    ┌─────────────────────┐
                                        │  Rollback       │    │  Status = Pending?  │
                                        │  Return 404     │    └──────────┬──────────┘
                                        └─────────────────┘               │
                                                              ┌───────────┴───────────┐
                                                              │ NO                    │ YES
                                                              ▼                       ▼
                                                    ┌─────────────────┐    ┌─────────────────────┐
                                                    │  Rollback       │    │  Action = REJECT?   │
                                                    │  Return 400     │    └──────────┬──────────┘
                                                    └─────────────────┘               │
                                                                          ┌───────────┴───────────┐
                                                                          │ YES                   │ NO (APPROVE)
                                                                          ▼                       ▼
                                                                ┌─────────────────────┐ ┌─────────────────────┐
                                                                │  Update Status to   │ │  Update Status to   │
                                                                │  REJECTED           │ │  APPROVED           │
                                                                └──────────┬──────────┘ └──────────┬──────────┘
                                                                           │                       │
                                                                           ▼                       ▼
                                                                ┌─────────────────────┐ ┌─────────────────────┐
                                                                │  Commit Transaction │ │  Update Leave       │
                                                                └──────────┬──────────┘ │  Balance (deduct)   │
                                                                           │            └──────────┬──────────┘
                                                                           │                       │
                                                                           │                       ▼
                                                                           │            ┌─────────────────────┐
                                                                           │            │  Commit Transaction │
                                                                           │            └──────────┬──────────┘
                                                                           │                       │
                                                                           └───────────┬───────────┘
                                                                                       │
                                                                                       ▼
                                                                            ┌─────────────────────┐
                                                                            │  Spawn Goroutine    │
                                                                            │  (Async Email)      │
                                                                            └──────────┬──────────┘
                                                                                       │
                                                                                       ▼
                                                                            ┌─────────────────────┐
                                                                            │  Fetch Employee     │
                                                                            │  Details            │
                                                                            └──────────┬──────────┘
                                                                                       │
                                                                                       ▼
                                                                            ┌─────────────────────┐
                                                                            │  Send Approval/     │
                                                                            │  Rejection Email    │
                                                                            └──────────┬──────────┘
                                                                                       │
                                                                                       ▼
                                                                            ┌─────────────────────┐
                                                                            │  Return 200         │
                                                                            │  Success Message    │
                                                                            └──────────┬──────────┘
                                                                                       │
                                                                                       ▼
                                                                                      END
```

---

### 4. Payroll Processing Activity Diagram

```
                              START
                                │
                                ▼
                  ┌──────────────────────────┐
                  │  Admin Initiates         │
                  │  Payroll Run             │
                  └────────────┬─────────────┘
                               │
                               ▼
                  ┌──────────────────────────┐
                  │  POST /api/payroll/run   │
                  │  {month, year}           │
                  └────────────┬─────────────┘
                               │
                               ▼
                  ┌──────────────────────────┐
                  │  Authenticate User       │
                  │  Check Role = ADMIN/     │
                  │  SUPERADMIN              │
                  └────────────┬─────────────┘
                               │
                               ▼
                        ┌──────────────┐
                        │  Authorized? │
                        └──────┬───────┘
                               │
                   ┌───────────┴───────────┐
                   │ NO                    │ YES
                   ▼                       ▼
          ┌─────────────────┐    ┌─────────────────────┐
          │  Return 403     │    │  Validate Month/Year│
          │  Forbidden      │    └──────────┬──────────┘
          └─────────────────┘               │
                                            ▼
                                 ┌─────────────────────┐
                                 │  Fetch Active       │
                                 │  Employees          │
                                 └──────────┬──────────┘
                                            │
                                            ▼
                                 ┌─────────────────────┐
                                 │  Fetch Working Days │
                                 │  from Settings      │
                                 └──────────┬──────────┘
                                            │
                                            ▼
                                 ┌─────────────────────┐
                                 │  For Each Employee: │
                                 │  ┌────────────────┐ │
                                 │  │ Calculate      │ │
                                 │  │ Absent Days    │ │
                                 │  │ (Approved      │ │
                                 │  │ Leaves)        │ │
                                 │  └────────┬───────┘ │
                                 │           │         │
                                 │           ▼         │
                                 │  ┌────────────────┐ │
                                 │  │ Calculate      │ │
                                 │  │ Deduction      │ │
                                 │  │ = Salary/Days  │ │
                                 │  │ × Absent       │ │
                                 │  └────────┬───────┘ │
                                 │           │         │
                                 │           ▼         │
                                 │  ┌────────────────┐ │
                                 │  │ Calculate      │ │
                                 │  │ Net Salary     │ │
                                 │  └────────────────┘ │
                                 └──────────┬──────────┘
                                            │
                                            ▼
                                 ┌─────────────────────┐
                                 │  Create Payroll Run │
                                 │  (Status: PREVIEW)  │
                                 └──────────┬──────────┘
                                            │
                                            ▼
                                 ┌─────────────────────┐
                                 │  Return 200         │
                                 │  - Payroll Run ID   │
                                 │  - Preview Data     │
                                 │  - Total Payroll    │
                                 └──────────┬──────────┘
                                            │
                                            ▼
                  ┌──────────────────────────────────────┐
                  │  Admin Reviews Preview               │
                  └────────────┬─────────────────────────┘
                               │
                               ▼
                  ┌──────────────────────────┐
                  │  POST /api/payroll/:id/  │
                  │  finalize                │
                  └────────────┬─────────────┘
                               │
                               ▼
                  ┌──────────────────────────┐
                  │  Start Transaction       │
                  └────────────┬─────────────┘
                               │
                               ▼
                  ┌──────────────────────────┐
                  │  For Each Employee:      │
                  │  ┌────────────────────┐  │
                  │  │ Create Payslip     │  │
                  │  │ Record             │  │
                  │  └────────────────────┘  │
                  └────────────┬─────────────┘
                               │
                               ▼
                  ┌──────────────────────────┐
                  │  Update Payroll Run      │
                  │  Status = FINALIZED      │
                  └────────────┬─────────────┘
                               │
                               ▼
                  ┌──────────────────────────┐
                  │  Commit Transaction      │
                  └────────────┬─────────────┘
                               │
                               ▼
                  ┌──────────────────────────┐
                  │  Return 200              │
                  │  - Payslip IDs           │
                  └────────────┬─────────────┘
                               │
                               ▼
                              END
```

---

## Sequence Diagrams

### 1. Employee Creation with Email Notification

```
Employee     Admin/HR      API Server      Database      Email Service
   │            │              │              │                │
   │            │              │              │                │
   │            ├─────────────►│              │                │
   │            │ POST /api/   │              │                │
   │            │ employee/    │              │                │
   │            │              │              │                │
   │            │              ├─────────────►│                │
   │            │              │ Validate     │                │
   │            │              │ Email        │                │
   │            │              │◄─────────────┤                │
   │            │              │              │                │
   │            │              ├─────────────►│                │
   │            │              │ Hash Password│                │
   │            │              │◄─────────────┤                │
   │            │              │              │                │
   │            │              ├─────────────►│                │
   │            │              │ INSERT       │                │
   │            │              │ Employee     │                │
   │            │              │◄─────────────┤                │
   │            │              │ Success      │                │
   │            │              │              │                │
   │            │◄─────────────┤              │                │
   │            │ 201 Created  │              │                │
   │            │              │              │                │
   │            │              ├──────────────┼───────────────►│
   │            │              │ Async Email  │                │
   │            │              │ (Goroutine)  │                │
   │            │              │              │                │
   │◄───────────┼──────────────┼──────────────┼────────────────┤
   │ Welcome    │              │              │                │
   │ Email with │              │              │                │
   │ Credentials│              │              │                │
   │            │              │              │                │
```

---

### 2. Leave Application with Notifications

```
Employee    Manager    API Server    Database    Email Service
   │           │           │             │              │
   │           │           │             │              │
   ├──────────────────────►│             │              │
   │ POST /api/leaves/     │             │              │
   │ apply                 │             │              │
   │                       │             │              │
   │                       ├────────────►│              │
   │                       │ Start TX    │              │
   │                       │             │              │
   │                       ├────────────►│              │
   │                       │ Fetch       │              │
   │                       │ Manager     │              │
   │                       │◄────────────┤              │
   │                       │             │              │
   │                       ├────────────►│              │
   │                       │ Calculate   │              │
   │                       │ Working Days│              │
   │                       │◄────────────┤              │
   │                       │             │              │
   │                       ├────────────►│              │
   │                       │ Check       │              │
   │                       │ Balance     │              │
   │                       │◄────────────┤              │
   │                       │             │              │
   │                       ├────────────►│              │
   │                       │ INSERT      │              │
   │                       │ Leave       │              │
   │                       │◄────────────┤              │
   │                       │             │              │
   │                       ├────────────►│              │
   │                       │ Commit TX   │              │
   │                       │◄────────────┤              │
   │                       │             │              │
   │◄──────────────────────┤             │              │
   │ 200 Success           │             │              │
   │                       │             │              │
   │                       ├─────────────┼─────────────►│
   │                       │ Async Email │              │
   │                       │ (Goroutine) │              │
   │                       │             │              │
   │           ◄───────────┼─────────────┼──────────────┤
   │           │ Leave     │             │              │
   │           │ Application             │              │
   │           │ Notification            │              │
   │           │           │             │              │
```

---

### 3. Leave Approval/Rejection Flow

```
Employee    Manager/Admin    API Server    Database    Email Service
   │             │               │             │              │
   │             │               │             │              │
   │             ├──────────────►│             │              │
   │             │ POST /api/    │             │              │
   │             │ leaves/:id/   │             │              │
   │             │ action        │             │              │
   │             │               │             │              │
   │             │               ├────────────►│              │
   │             │               │ Start TX    │              │
   │             │               │             │              │
   │             │               ├────────────►│              │
   │             │               │ Fetch Leave │              │
   │             │               │ (with lock) │              │
   │             │               │◄────────────┤              │
   │             │               │             │              │
   │             │               ├────────────►│              │
   │             │               │ UPDATE      │              │
   │             │               │ Status      │              │
   │             │               │◄────────────┤              │
   │             │               │             │              │
   │             │               ├────────────►│              │
   │             │               │ UPDATE      │              │
   │             │               │ Balance     │              │
   │             │               │ (if approved)              │
   │             │               │◄────────────┤              │
   │             │               │             │              │
   │             │               ├────────────►│              │
   │             │               │ Commit TX   │              │
   │             │               │◄────────────┤              │
   │             │               │             │              │
   │             │◄──────────────┤             │              │
   │             │ 200 Success   │             │              │
   │             │               │             │              │
   │             │               ├─────────────┼─────────────►│
   │             │               │ Async Email │              │
   │             │               │ (Goroutine) │              │
   │             │               │             │              │
   │◄────────────┼───────────────┼─────────────┼──────────────┤
   │ Approval/   │               │             │              │
   │ Rejection   │               │             │              │
   │ Email       │               │             │              │
   │             │               │             │              │
```

---

## Database Relationships Summary

### One-to-Many Relationships

1. **Tbl_Role → Tbl_Employee**
   - One role can have many employees
   - FK: `Tbl_Employee.role_id` → `Tbl_Role.id`

2. **Tbl_Employee → Tbl_Employee (Self-Referencing)**
   - One manager can manage many employees
   - FK: `Tbl_Employee.manager_id` → `Tbl_Employee.id`

3. **Tbl_Employee → Tbl_Leave**
   - One employee can have many leave applications
   - FK: `Tbl_Leave.employee_id` → `Tbl_Employee.id`

4. **Tbl_Leave_type → Tbl_Leave**
   - One leave type can have many leave applications
   - FK: `Tbl_Leave.leave_type_id` → `Tbl_Leave_type.id`

5. **Tbl_Employee → Tbl_Leave_balance**
   - One employee can have many leave balances (per year, per type)
   - FK: `Tbl_Leave_balance.employee_id` → `Tbl_Employee.id`

6. **Tbl_Leave_type → Tbl_Leave_balance**
   - One leave type can have many balance records
   - FK: `Tbl_Leave_balance.leave_type_id` → `Tbl_Leave_type.id`

7. **Tbl_Employee → Tbl_Leave_adjustment**
   - One employee can have many leave adjustments
   - FK: `Tbl_Leave_adjustment.employee_id` → `Tbl_Employee.id`

8. **Tbl_Payroll_run → Tbl_Payslip**
   - One payroll run can have many payslips
   - FK: `Tbl_Payslip.payroll_run_id` → `Tbl_Payroll_run.id`

9. **Tbl_Employee → Tbl_Payslip**
   - One employee can have many payslips
   - FK: `Tbl_Payslip.employee_id` → `Tbl_Employee.id`

10. **Tbl_Employee → Tbl_Audit**
    - One employee can perform many audited actions
    - FK: `Tbl_Audit.actor_id` → `Tbl_Employee.id`

---

## Indexes Recommendations

For optimal performance, consider adding these indexes:

```sql
-- Employee lookups
CREATE INDEX idx_employee_email ON Tbl_Employee(email);
CREATE INDEX idx_employee_role ON Tbl_Employee(role_id);
CREATE INDEX idx_employee_manager ON Tbl_Employee(manager_id);
CREATE INDEX idx_employee_status ON Tbl_Employee(status);

-- Leave queries
CREATE INDEX idx_leave_employee ON Tbl_Leave(employee_id);
CREATE INDEX idx_leave_status ON Tbl_Leave(status);
CREATE INDEX idx_leave_dates ON Tbl_Leave(start_date, end_date);

-- Leave balance lookups
CREATE INDEX idx_balance_employee_year ON Tbl_Leave_balance(employee_id, year);
CREATE INDEX idx_balance_type ON Tbl_Leave_balance(leave_type_id);

-- Payroll queries
CREATE INDEX idx_payslip_run ON Tbl_Payslip(payroll_run_id);
CREATE INDEX idx_payslip_employee ON Tbl_Payslip(employee_id);
CREATE INDEX idx_payroll_month_year ON Tbl_Payroll_run(month, year);

-- Audit trail
CREATE INDEX idx_audit_actor ON Tbl_Audit(actor_id);
CREATE INDEX idx_audit_entity ON Tbl_Audit(entity, entity_id);
CREATE INDEX idx_audit_created ON Tbl_Audit(created_at);

-- Holiday lookups
CREATE INDEX idx_holiday_date ON Tbl_Holiday(date);
```

---

## Data Integrity Constraints

### Foreign Key Constraints
All foreign keys are enforced with `REFERENCES` clause ensuring referential integrity.

### Unique Constraints
- `Tbl_Employee.email` - Ensures unique email addresses
- `Tbl_Role.type` - Ensures unique role names
- `Tbl_Holiday.date` - Ensures no duplicate holidays on same date

### Check Constraints (Recommended)
```sql
-- Ensure valid status values
ALTER TABLE Tbl_Employee 
ADD CONSTRAINT chk_employee_status 
CHECK (status IN ('active', 'inactive'));

-- Ensure valid leave status
ALTER TABLE Tbl_Leave 
ADD CONSTRAINT chk_leave_status 
CHECK (status IN ('Pending', 'MANAGER_APPROVED', 'MANAGER_REJECTED', 'APPROVED', 'REJECTED', 'CANCELLED', 'WITHDRAWAL_PENDING', 'WITHDRAWN'));

-- Ensure valid payroll status
ALTER TABLE Tbl_Payroll_run 
ADD CONSTRAINT chk_payroll_status 
CHECK (status IN ('PREVIEW', 'FINALIZED'));

-- Ensure positive salary
ALTER TABLE Tbl_Employee 
ADD CONSTRAINT chk_positive_salary 
CHECK (salary >= 0);

-- Ensure valid month
ALTER TABLE Tbl_Payroll_run 
ADD CONSTRAINT chk_valid_month 
CHECK (month BETWEEN 1 AND 12);

-- Ensure end_date >= start_date
ALTER TABLE Tbl_Leave 
ADD CONSTRAINT chk_valid_date_range 
CHECK (end_date >= start_date);
```

---

## Database Statistics

Based on the schema:

- **Total Tables**: 11
- **Total Relationships**: 15+ foreign keys
- **UUID Primary Keys**: 8 tables
- **Serial Primary Keys**: 3 tables
- **Self-Referencing Tables**: 1 (Tbl_Employee)
- **Audit Enabled**: Yes (Tbl_Audit)
- **Soft Delete Enabled**: Yes (Tbl_Employee.deleted_at)

---

**Last Updated**: November 26, 2024  
**Database**: PostgreSQL  
**Schema Version**: 1.0
