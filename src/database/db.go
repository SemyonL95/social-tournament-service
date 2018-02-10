package database

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

const (
	user     = "postgres"
	dbname   = "postgres"
	password = "mypass"
	host     = "database"
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

	return databaseConn, nil
}

func (db *DB) Truncate() error {
	return retryTransact(db.conn, db.txTruncate(), 5)
}

func (db *DB) txTruncate() func(*sqlx.Tx) error {
	return func(tx *sqlx.Tx) error {
		_, err := tx.Exec("TRUNCATE users CASCADE")
		if err != nil {
			return err
		}

		_, err = tx.Exec("TRUNCATE tournaments CASCADE")
		if err != nil {
			return nil
		}
		return nil
	}
}
