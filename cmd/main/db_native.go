//go:build !cgo_sqlite

package main

import (
	"database/sql"
	_ "modernc.org/sqlite"
)

func initDB(dataSource string) (*sql.DB, error) {
	return sql.Open("sqlite", dataSource)
}
