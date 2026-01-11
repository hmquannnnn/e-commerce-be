-- Create users table
CREATE TABLE
IF NOT EXISTS users
(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    avatar_url TEXT,
    date_of_birth DATE,
    gender VARCHAR(10) NOT NULL CHECK(gender IN('male', 'female', 'other')),
    role VARCHAR(20) NOT NULL DEFAULT 'user' CHECK(role IN('user', 'seller', 'admin')),
    
    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
