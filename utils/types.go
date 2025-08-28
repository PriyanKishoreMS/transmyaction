package utils

import "time"

type Transaction struct {
	ID            int       `json:"id,omitempty" db:"id"`
	UserEmail     string    `json:"userEmail" db:"user_email"`
	Amount        float64   `json:"amount" db:"amount"`
	AccountNumber string    `json:"accountNumber" db:"account_number"`
	TxnMethod     string    `json:"txnMethod" db:"txn_method"`
	TxnMode       string    `json:"txnMode" db:"txn_mode"`
	TxnType       string    `json:"txnType" db:"txn_type"`
	TxnRef        string    `json:"txnRef" db:"txn_ref"`
	CounterParty  string    `json:"counterParty" db:"counter_party"`
	TxnInfo       string    `json:"txnInfo" db:"txn_info"`
	TxnDatetime   time.Time `json:"txnDatetime" db:"txn_datetime"`
	CreatedTime   time.Time `json:"createdTime,omitempty" db:"created_time"`
}
