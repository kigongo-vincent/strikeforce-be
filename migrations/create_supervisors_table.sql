-- Migration: Create supervisors table
-- This table links users to departments for supervisor role
-- Run this manually if AutoMigrate hasn't created the table yet

CREATE TABLE IF NOT EXISTS supervisors (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    user_id INTEGER NOT NULL,
    department_id INTEGER NOT NULL,
    CONSTRAINT fk_supervisors_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT fk_supervisors_department FOREIGN KEY (department_id) REFERENCES departments (id) ON DELETE CASCADE,
    CONSTRAINT unique_user_department UNIQUE (user_id, department_id)
);

CREATE INDEX IF NOT EXISTS idx_supervisors_user_id ON supervisors (user_id);

CREATE INDEX IF NOT EXISTS idx_supervisors_department_id ON supervisors (department_id);

CREATE INDEX IF NOT EXISTS idx_supervisors_deleted_at ON supervisors (deleted_at);