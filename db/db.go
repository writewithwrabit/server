package db

import (
	"database/sql"
	"fmt"
)

func LogAndQuery(db *sql.DB, query string, args ...interface{}) *sql.Rows {
	fmt.Println(query)
	fmt.Println(args...)

	res, err := db.Query(query, args...)
	if err != nil {
		panic(err)
	}

	return res
}

func LogAndQueryRow(db *sql.DB, query string, args ...interface{}) *sql.Row {
	fmt.Println(query)
	fmt.Println(args...)

	return db.QueryRow(query, args...)
}

func LogAndExec(db *sql.DB, query string, args ...interface{}) sql.Result {
	fmt.Println(query)
	fmt.Println(args...)

	res, err := db.Exec(query, args...)
	if err != nil {
		panic(err)
	}

	return res
}
