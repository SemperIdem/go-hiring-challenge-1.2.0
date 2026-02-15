CREATE TABLE IF NOT EXISTS product_categories (
    id SERIAL PRIMARY KEY,
    code VARCHAR(32) UNIQUE NOT NULL,
    name VARCHAR(256) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

ALTER TABLE products
ADD COLUMN IF NOT EXISTS category_id INTEGER REFERENCES product_categories(id);

CREATE INDEX IF NOT EXISTS idx_products_category_id ON products(category_id);
