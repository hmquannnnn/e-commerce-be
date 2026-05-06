ALTER TABLE payments DROP CONSTRAINT IF EXISTS payments_payment_method_check;
ALTER TABLE payments ADD CONSTRAINT payments_payment_method_check
    CHECK (payment_method IN ('card', 'wallet', 'atm'));

ALTER TABLE payments DROP CONSTRAINT IF EXISTS payments_provider_check;
ALTER TABLE payments ADD CONSTRAINT payments_provider_check
    CHECK (provider IN ('stripe', 'vnpay'));

ALTER TABLE payments DROP COLUMN IF EXISTS payos_order_code;

DROP SEQUENCE IF EXISTS payos_order_code_seq;
