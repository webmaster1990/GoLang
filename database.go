package main

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var pg *sqlx.DB

func connectPostgres() {
	if pg != nil {
		// already connected
		return
	}
	log.Println("Connect to PostgreSQL")
	connString := fmt.Sprintf("dbname=%s user=%s password=%s host=%s sslmode=disable search_path=%s",
		DB_NAME, DB_USER, DB_PASS, DB_HOST, DB_SCHEMA)
	pg = sqlx.MustConnect("postgres", connString)
	pg.SetMaxIdleConns(1)
	pg.SetMaxOpenConns(8)
	log.Println("... Connected to PostgreSQL")
}
