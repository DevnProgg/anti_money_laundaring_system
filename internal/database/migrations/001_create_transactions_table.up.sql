
-- DELIVERABLE 1: SQL CREATE TABLE statement
CREATE TABLE transactions (
    transaction_id UUID PRIMARY KEY,
    account_id VARCHAR(255) NOT NULL,
    amount DECIMAL(18, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    "timestamp" TIMESTAMP WITH TIME ZONE NOT NULL,
    source_country VARCHAR(2) NOT NULL,
    destination_country VARCHAR(2) NOT NULL,
    transaction_type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL,
    CONSTRAINT positive_amount CHECK (amount > 0)
);

-- DELIVERABLE 2: CREATE INDEX statements
CREATE INDEX idx_transactions_account_id ON transactions(account_id);
CREATE INDEX idx_transactions_timestamp ON transactions("timestamp");
CREATE INDEX idx_transactions_amount ON transactions(amount);
