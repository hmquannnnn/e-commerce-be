CREATE TABLE IF NOT EXISTS payment_attempts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    payment_id UUID NOT NULL REFERENCES payments(id) ON DELETE CASCADE,
    attempt_no INT NOT NULL CHECK (attempt_no > 0),
    provider_session_id VARCHAR(255),
    checkout_url TEXT,
    status VARCHAR(30) NOT NULL,
    raw_request JSONB,
    raw_response JSONB,
    error_code VARCHAR(100),
    error_message TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payment_attempts_payment_id ON payment_attempts(payment_id);
