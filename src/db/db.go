package db

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"

	"../models"
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

type NotFoundError struct {
	text string
}

type ForbiddenError struct {
	text string
}

func (err *ForbiddenError) Error() string {
	return fmt.Sprintf("%s %s", err.text, "Forbidden")
}

func (err *NotFoundError) Error() string {
	return fmt.Sprintf("%s %s", err.text, "Not Found")
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

func (db *DB) FundOrCreateUser(username string, credits float64) (*models.User, error) {
	sql := `INSERT INTO users (username, credits) VALUES ($1, $2) 
			ON CONFLICT (username) DO UPDATE SET credits = $2 RETURNING *;`

	stmt, err := db.conn.Prepare(sql)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	defer stmt.Close()

	user := models.User{}
	err = stmt.QueryRow(username, credits).Scan(&user.Id, &user.Username, &user.Credits)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	return &user, nil
}

func (db *DB) TakePointsFromUser(username string, credits float64) (*models.User, error, *NotFoundError, *ForbiddenError) {
	tx := db.conn.MustBegin()
	user := models.User{}

	err := tx.Get(&user, `SELECT * FROM users WHERE username = $1 FOR UPDATE`, username)
	if err != nil {
		errMsg := fmt.Sprintf("User With playerID - %s", username)
		return nil, nil, &NotFoundError{errMsg}, nil
	}

	if (user.Credits - credits) < 0 {
		return nil, nil, nil, &ForbiddenError{"User don't have enough points"}
	}

	user.Credits = user.Credits - credits
	tx.NamedExec("UPDATE users SET credits = :credits WHERE username = :username", &user)
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return nil, err, nil, nil
	}

	return &user, nil, nil, nil
}

func (db *DB) GetByUsername(username string) (*models.User, *NotFoundError) {
	user := models.User{}
	err := db.conn.Get(&user, `SELECT * FROM users WHERE username = $1 FOR UPDATE`, username)

	if err != nil {
		errMsg := fmt.Sprintf("User With playerID - %s", username)
		return nil, &NotFoundError{errMsg}
	}

	return &user, nil
}
