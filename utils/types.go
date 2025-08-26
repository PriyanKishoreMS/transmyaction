package utils

import "time"

type Transaction struct {
	UserEmail     string
	Amount        float64
	AccountNumber string
	TxnMethod     string
	TxnMode       string
	TxnType       string
	TxnRef        string
	CounterParty  string
	TxnInfo       string
	TxnDatetime   time.Time
}
