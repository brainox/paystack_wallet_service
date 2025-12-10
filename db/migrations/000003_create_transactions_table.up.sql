CREATE TYPE transaction_type AS ENUM ('deposit', 'transfer', 'credit', 'debit');
CREATE TYPE transaction_status AS ENUM ('pending', 'success', 'failed');

CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    type transaction_type NOT NULL,
    amount DECIMAL(20, 2) NOT NULL CHECK (amount > 0),
    status transaction_status DEFAULT 'pending' NOT NULL,
    reference VARCHAR(255) UNIQUE,
    paystack_reference VARCHAR(255),
    recipient_wallet_id UUID REFERENCES wallets(id),
    recipient_user_id UUID REFERENCES users(id),
    description TEXT,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_transactions_user_id ON transactions(user_id);
CREATE INDEX idx_transactions_wallet_id ON transactions(wallet_id);
CREATE INDEX idx_transactions_reference ON transactions(reference);
CREATE INDEX idx_transactions_paystack_reference ON transactions(paystack_reference);
CREATE INDEX idx_transactions_status ON transactions(status);
CREATE INDEX idx_transactions_type ON transactions(type);
CREATE INDEX idx_transactions_created_at ON transactions(created_at DESC);
