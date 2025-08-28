package data

import (
	"fmt"
	"time"

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

func (t TxnModel) GetTransactions(mail string, interval string, year, month int, from, to *time.Time) ([]utils.Transaction, error) {
	now := time.Now()
	var start, end time.Time

	switch {
	// custom date range
	case from != nil && to != nil:
		start, end = *from, *to

	// last 7 days
	case interval == "7d":
		start, end = now.AddDate(0, 0, -7), now

	// specific month navigation
	case interval == "month":
		if year > 0 && month > 0 {
			loc := now.Location()
			start = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, loc)
			end = start.AddDate(0, 1, 0).Add(-time.Nanosecond) // last nanosecond of the month
		} else {
			// fallback: current month
			start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
			end = now
		}

	default:
		return nil, fmt.Errorf("unsupported interval: %s", interval)
	}

	query := `
        SELECT id, user_email, amount, account_number, txn_method, txn_mode, txn_type,
               txn_ref, counter_party, txn_info, txn_datetime, created_at
        FROM transactions
        WHERE user_email = ?
          AND txn_datetime BETWEEN ? AND ?
        ORDER BY txn_datetime DESC
    `

	rows, err := t.DB.Query(query, mail, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txns []utils.Transaction
	for rows.Next() {
		var txn utils.Transaction
		if err := rows.Scan(
			&txn.ID, &txn.UserEmail, &txn.Amount, &txn.AccountNumber,
			&txn.TxnMethod, &txn.TxnMode, &txn.TxnType, &txn.TxnRef,
			&txn.CounterParty, &txn.TxnInfo, &txn.TxnDatetime, &txn.CreatedTime,
		); err != nil {
			return nil, err
		}
		txns = append(txns, txn)
	}

	return txns, rows.Err()
}
