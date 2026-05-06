CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id VARCHAR(100) NOT NULL,
    user_id UUID NOT NULL,
    provider VARCHAR(20) NOT NULL CHECK (provider IN ('stripe', 'vnpay')),
    payment_method VARCHAR(20) NOT NULL CHECK (payment_method IN ('card', 'wallet', 'atm')),
    amount BIGINT NOT NULL CHECK (amount > 0),
    currency VARCHAR(10) NOT NULL DEFAULT 'VND',
    status VARCHAR(30) NOT NULL DEFAULT 'pending'
      CHECK (status IN ('pending', 'requires_action', 'processing', 'succeeded', 'failed', 'canceled', 'expired', 'refunded_partial', 'refunded_full')),
    provider_payment_id VARCHAR(255),
    provider_transaction_ref VARCHAR(255),
    return_url TEXT,
    cancel_url TEXT,
    checkout_url TEXT,
    expires_at TIMESTAMP,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payments_order_id ON payments(order_id);
CREATE INDEX IF NOT EXISTS idx_payments_user_id ON payments(user_id);
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);
CREATE INDEX IF NOT EXISTS idx_payments_provider_payment_id ON payments(provider_payment_id);
CREATE INDEX IF NOT EXISTS idx_payments_provider_transaction_ref ON payments(provider_transaction_ref);
