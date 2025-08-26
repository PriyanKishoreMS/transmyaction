package data

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/priyankishorems/transmyaction/utils"
)

type TxnModel struct {
	DB *sqlx.DB
}

func (t TxnModel) SaveTransactions(allTxn []utils.Transaction) error {
	tx, err := t.DB.Beginx()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	stmt, err := tx.Preparex(`
        INSERT OR IGNORE INTO transactions (
            user_email, amount, account_number, txn_method, txn_mode, txn_type,
            txn_ref, counter_party, txn_info, txn_datetime
        )
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("prepare stmt: %w", err)
	}
	defer stmt.Close()

	for _, txn := range allTxn {
		_, err := stmt.Exec(
			txn.UserEmail,
			txn.Amount,
			txn.AccountNumber,
			txn.TxnMethod,
			txn.TxnMode,
			txn.TxnType,
			txn.TxnRef,
			txn.CounterParty,
			txn.TxnInfo,
			txn.TxnDatetime,
		)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("insert txn: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}
