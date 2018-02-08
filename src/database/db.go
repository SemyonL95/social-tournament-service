package database

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"

	"github.com/SemyonL95/social-tournament-service/src/models"
	"errors"
)

const (
	user     = "postgres"
	dbname   = "postgres"
	password = "mypass"
	host     = "database"
	port     = "5432"
)

var ErrNotEnoughMoney = errors.New("not enough money")

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

func (db *DB) FundOrCreateUser(username string, credits float64) error {
	sql := `INSERT INTO users (username, credits) VALUES ($1, $2) 
			ON CONFLICT (username) DO UPDATE SET credits = $2;`

	_, err := db.conn.Exec(sql, username, credits)

	if err != nil {
		log.Println(err.Error())
		return err
	}

	return nil
}

func (db *DB) TakePointsFromUser(username string, credits float64) error {
	tx, err := db.conn.Beginx()
	if err != nil {
		log.Println(err)
		return err
	}
	user := models.User{}

	err = tx.Get(&user, `SELECT * FROM users WHERE username = $1 FOR UPDATE`, username)
	if err != nil {
		return err
	}

	if (user.Credits - credits) < 0 {
		return ErrNotEnoughMoney
	}

	user.Credits = user.Credits - credits
	tx.NamedExec(`UPDATE users SET credits = :credits WHERE username = :username`, &user)
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (db *DB) CreateTournament(id int, deposit float64) error {
	sql := `INSERT INTO tournaments (id, deposit) VALUES ($1, $2)`

	_, err := db.conn.Exec(sql, id, deposit)
	if err != nil {
		return err
	}

	return nil
}
