ALTER TABLE orders
    DROP COLUMN IF EXISTS shipping_phone,
    DROP COLUMN IF EXISTS shipping_address;
