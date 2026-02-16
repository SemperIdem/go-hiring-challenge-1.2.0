INSERT INTO product_categories (code, name) VALUES
('CLOTHING', 'Clothing'),
('SHOES', 'Shoes'),
('ACCESSORIES', 'Accessories');

UPDATE products
SET category_id = (SELECT id FROM product_categories WHERE code = 'CLOTHING')
WHERE code IN ('PROD001', 'PROD004', 'PROD007');

UPDATE products
SET category_id = (SELECT id FROM product_categories WHERE code = 'SHOES')
WHERE code IN ('PROD002', 'PROD006');

UPDATE products
SET category_id = (SELECT id FROM product_categories WHERE code = 'ACCESSORIES')
WHERE code IN ('PROD003', 'PROD005', 'PROD008');

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM products WHERE category_id IS NULL) THEN
        RAISE NOTICE 'Skipping NOT NULL on products.category_id because unmapped products still exist';
    ELSE
        ALTER TABLE products
        ALTER COLUMN category_id SET NOT NULL;
    END IF;
END $$;
