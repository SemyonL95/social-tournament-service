package db

import (
	"fmt"
	"github.com/jmoiron/sqlx"
)

const (
	user     = "postgres"
	dbname   = "postgres"
	password = "mypass"
	host     = "db"
	port     = "5432"
)

type DB struct {
	conn *sqlx.DB
}

func InitDatabaseConn() (*DB, error) {
	connInfo := fmt.Sprintf(
		"user=%s dbname=%s password=%s host=%s port=%s sslmode=disable",
		user,
		dbname,
		password,
		host,
		port,
	)

	db, err := sqlx.Open("postgres", connInfo)

	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	databaseConn := &DB{db}

	fmt.Println("Database connections setup successfuly")

	return databaseConn, nil
}

func (db *DB) Testdb() {
	fmt.Println("DBTESTED")
}
