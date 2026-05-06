CREATE TABLE IF NOT EXISTS payment_webhook_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    provider VARCHAR(20) NOT NULL CHECK (provider IN ('stripe', 'vnpay')),
    event_id VARCHAR(255) NOT NULL,
    signature_valid BOOLEAN NOT NULL DEFAULT false,
    payload JSONB NOT NULL,
    processed_at TIMESTAMP NULL,
    process_status VARCHAR(20) NOT NULL DEFAULT 'new'
      CHECK (process_status IN ('new', 'processed', 'ignored', 'failed')),
    error TEXT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (provider, event_id)
);
