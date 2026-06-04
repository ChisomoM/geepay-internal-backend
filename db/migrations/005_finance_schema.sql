-- Finance domain schema

CREATE TABLE IF NOT EXISTS licenses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companys(id),
    name VARCHAR(255) NOT NULL,
    provider VARCHAR(255),
    renewal_date TIMESTAMP,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_licenses_company_id ON licenses(company_id);

CREATE TABLE IF NOT EXISTS budget_and_licenses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companys(id),
    name VARCHAR(255) NOT NULL,
    renewal_date TIMESTAMP,
    actual_amount NUMERIC,
    purchase_frequency VARCHAR(255),
    license_id UUID REFERENCES licenses(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_budget_and_licenses_company_id ON budget_and_licenses(company_id);
CREATE INDEX IF NOT EXISTS idx_budget_and_licenses_license_id ON budget_and_licenses(license_id);

CREATE TABLE IF NOT EXISTS statutories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companys(id),
    name VARCHAR(255) NOT NULL,
    due_date TIMESTAMP,
    amount NUMERIC,
    is_paid BOOLEAN DEFAULT false,
    paid_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_statutories_company_id ON statutories(company_id);

CREATE TABLE IF NOT EXISTS salary_advances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companys(id),
    user_id UUID NOT NULL REFERENCES users(id),
    amount NUMERIC NOT NULL,
    status VARCHAR(100),
    deduction_months INTEGER DEFAULT 0,
    approved_by UUID REFERENCES users(id),
    approved_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_salary_advances_company_id ON salary_advances(company_id);
CREATE INDEX IF NOT EXISTS idx_salary_advances_user_id ON salary_advances(user_id);
