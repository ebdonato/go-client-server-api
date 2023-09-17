package database

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func OpenDatabase() *sql.DB {
	var err error

	db, err = sql.Open("sqlite3", "./db.sqlite")
	if err != nil {
		panic(err)
	}

	stmt := `
        CREATE TABLE IF NOT EXISTS exchanges(
            id INTEGER PRIMARY KEY,
            code TEXT,
            code_in TEXT,
            name TEXT,
            high TEXT,
            low TEXT,
            var_bid TEXT,
            pct_change TEXT,
            bid TEXT,
            ask TEXT,
            timestamp TEXT,
            create_date TEXT,
			persist_date DATETIME DEFAULT CURRENT_TIMESTAMP
        );
    `
	_, err = db.Exec(stmt)

	if err != nil {
		panic(err)
	}

	return db
}

func GetDatabase() *sql.DB {
	return db
}
