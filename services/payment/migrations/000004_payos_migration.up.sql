CREATE SEQUENCE IF NOT EXISTS payos_order_code_seq START 100001;

ALTER TABLE payments ADD COLUMN IF NOT EXISTS payos_order_code BIGINT UNIQUE;

ALTER TABLE payments DROP CONSTRAINT IF EXISTS payments_provider_check;
ALTER TABLE payments ADD CONSTRAINT payments_provider_check
    CHECK (provider IN ('payos'));

ALTER TABLE payments DROP CONSTRAINT IF EXISTS payments_payment_method_check;
ALTER TABLE payments ADD CONSTRAINT payments_payment_method_check
    CHECK (payment_method IN ('qr_code'));
