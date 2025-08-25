package data

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/priyankishorems/transmyaction/utils"
	_ "modernc.org/sqlite"
)

type SQLiteDB struct {
	Database string
}
type dbInfo struct {
	Seq  int    `db:"seq"`
	Name string `db:"name"`
	File string `db:"file"`
}

func (m SQLiteDB) Open() (*sqlx.DB, error) {

	s := SQLiteDB{
		Database: utils.DBName,
	}

	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(ON)", s.Database)

	db, err := sqlx.Connect("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to sqlite: %w", err)
	}

	var info dbInfo
	err = db.Get(&info, "PRAGMA database_list;")
	if err != nil {
		return nil, fmt.Errorf("failed to query database_list: %w", err)
	}

	log.Printf("Connected to SQLite DB file: %s", info.File)

	return db, nil
}
