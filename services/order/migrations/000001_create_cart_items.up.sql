CREATE TABLE IF NOT EXISTS cart_items (
    id           BIGSERIAL PRIMARY KEY,
    user_id      UUID NOT NULL,
    product_id   UUID NOT NULL,
    product_name VARCHAR(255) NOT NULL,
    unit_price   DECIMAL(12,2) NOT NULL CHECK (unit_price >= 0),
    image_url    TEXT,
    quantity     INT NOT NULL DEFAULT 1 CHECK (quantity > 0),
    updated_at   TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, product_id)
);

CREATE INDEX IF NOT EXISTS idx_cart_items_user_id ON cart_items(user_id);
