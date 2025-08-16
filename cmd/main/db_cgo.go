//go:build cgo_sqlite

package main

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func initDB(dataSource string) (*sql.DB, error) {
	return sql.Open("sqlite3", dataSource)
}
