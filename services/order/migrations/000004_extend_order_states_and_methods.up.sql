-- Extend orders state machine:
--   PROCESSING is replaced by DELIVERING (đang giao) and DELIVERED (đã giao xong).
--   payment_method now also accepts 'CARD' (Stripe / international card).
--
-- Existing rows with status='PROCESSING' are migrated to 'DELIVERING' so the
-- new CHECK constraint passes. There is no historical CARD data to backfill.

UPDATE orders SET status = 'DELIVERING' WHERE status = 'PROCESSING';

ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_status_check;
ALTER TABLE orders ADD CONSTRAINT orders_status_check
    CHECK (status IN ('PENDING', 'PAID', 'DELIVERING', 'DELIVERED', 'CANCELLED'));

ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_payment_method_check;
ALTER TABLE orders ADD CONSTRAINT orders_payment_method_check
    CHECK (payment_method IN ('VNPAY', 'CASH', 'CARD'));
