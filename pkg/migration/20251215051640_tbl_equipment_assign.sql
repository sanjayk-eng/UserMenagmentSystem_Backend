-- +goose Up
-- +goose StatementBegin

-- ===============================
-- Equipment Category
-- ===============================
CREATE TABLE IF NOT EXISTS tbl_equipment_category (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),  -- auto-generated UUID
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT DEFAULT '',                    -- single quotes

    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now()
);

-- ===============================
-- Equipment Master
-- ===============================
CREATE TABLE IF NOT EXISTS tbl_equipment (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),  -- auto-generated UUID
    name VARCHAR(100) NOT NULL,

    category_id UUID NOT NULL
        REFERENCES tbl_equipment_category(id)
        ON DELETE RESTRICT,

    ownership VARCHAR(20) NOT NULL
        CHECK (ownership IN ('COMPANY','SELF'))
        DEFAULT 'COMPANY',

    is_shared BOOLEAN NOT NULL DEFAULT FALSE,
    total_quantity INT NOT NULL CHECK (total_quantity >= 0),

    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now()
);

-- ===============================
-- Equipment Assignment
-- ===============================
CREATE TABLE IF NOT EXISTS tbl_equipment_assignment (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),  -- auto-generated UUID

    equipment_id UUID NOT NULL
        REFERENCES tbl_equipment(id)
        ON DELETE RESTRICT,

    employee_id UUID NOT NULL
        REFERENCES tbl_employee(id)                 -- lowercase table
        ON DELETE RESTRICT,

    quantity INT NOT NULL DEFAULT 1 CHECK (quantity > 0),

    assigned_at TIMESTAMP NOT NULL DEFAULT now(),
    returned_at TIMESTAMP
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tbl_equipment_assignment;
DROP TABLE IF EXISTS tbl_equipment;
DROP TABLE IF EXISTS tbl_equipment_category;
-- +goose StatementEnd
