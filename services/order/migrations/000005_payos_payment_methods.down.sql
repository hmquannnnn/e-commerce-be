ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_payment_method_check;
ALTER TABLE orders ADD CONSTRAINT orders_payment_method_check
    CHECK (payment_method IN ('VNPAY', 'CASH', 'CARD'));

UPDATE orders SET payment_method = 'VNPAY' WHERE payment_method = 'QR_CODE';
