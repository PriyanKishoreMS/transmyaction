-- +goose Up
-- +goose StatementBegin
CREATE TABLE transactions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_email TEXT NOT NULL,
    amount REAL NOT NULL,
    account_number TEXT,
    txn_method TEXT, -- upi/neft
    txn_mode TEXT, -- p2a
    txn_type TEXT, -- credit/debit        
    txn_ref TEXT, -- ref no.
    counter_party TEXT,
    txn_info TEXT,    
    txn_datetime DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_email) REFERENCES gmail_tokens(user_email)
);
CREATE UNIQUE INDEX uniq_txn ON transactions(user_email, txn_ref);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

