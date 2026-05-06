-- Revert: collapse DELIVERING/DELIVERED back to PROCESSING and drop CARD.
-- Any CARD orders are forcibly converted to VNPAY so the constraint passes.

UPDATE orders SET status = 'PROCESSING' WHERE status IN ('DELIVERING', 'DELIVERED');
UPDATE orders SET payment_method = 'VNPAY' WHERE payment_method = 'CARD';

ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_status_check;
ALTER TABLE orders ADD CONSTRAINT orders_status_check
    CHECK (status IN ('PENDING', 'PAID', 'PROCESSING', 'CANCELLED'));

ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_payment_method_check;
ALTER TABLE orders ADD CONSTRAINT orders_payment_method_check
    CHECK (payment_method IN ('VNPAY', 'CASH'));
