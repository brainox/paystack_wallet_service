CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    wallet_number VARCHAR(13) UNIQUE NOT NULL,
    balance DECIMAL(20, 2) DEFAULT 0.00 NOT NULL CHECK (balance >= 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT unique_user_wallet UNIQUE(user_id)
);

CREATE INDEX idx_wallets_user_id ON wallets(user_id);
CREATE INDEX idx_wallets_wallet_number ON wallets(wallet_number);

-- Function to generate wallet number
CREATE OR REPLACE FUNCTION generate_wallet_number() RETURNS VARCHAR(13) AS $$
DECLARE
    new_wallet_number VARCHAR(13);
    done BOOLEAN;
BEGIN
    done := false;
    WHILE NOT done LOOP
        new_wallet_number := LPAD(FLOOR(RANDOM() * 10000000000000)::TEXT, 13, '0');
        done := NOT EXISTS(SELECT 1 FROM wallets WHERE wallet_number = new_wallet_number);
    END LOOP;
    RETURN new_wallet_number;
END;
$$ LANGUAGE plpgsql;
