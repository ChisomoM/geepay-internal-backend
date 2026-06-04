-- Inventory domain schema

CREATE TABLE IF NOT EXISTS inventory_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companys(id),
    name VARCHAR(255) NOT NULL,
    low_stock_threshold INTEGER DEFAULT 0,
    out_of_stock_threshold INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_inventory_categories_company_id ON inventory_categories(company_id);

CREATE TABLE IF NOT EXISTS inventory (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companys(id),
    category_id UUID REFERENCES inventory_categories(id),
    name VARCHAR(255) NOT NULL,
    quantity INTEGER DEFAULT 0,
    status VARCHAR(100),
    assigned_to UUID,
    assigned_type VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_inventory_company_id ON inventory(company_id);
CREATE INDEX IF NOT EXISTS idx_inventory_category_id ON inventory(category_id);
