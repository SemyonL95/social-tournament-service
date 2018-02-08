package db

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"

	"../models"
	"../utils"
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

func (db *DB) TakePointsFromUser(username string, credits float64) (*models.User, error) {
	tx := db.conn.MustBegin()
	user := models.User{}

	err := tx.Get(&user, `SELECT * FROM users WHERE username = $1 FOR UPDATE`, username)
	if err != nil {
		errMsg := fmt.Sprintf("User With playerID - %s", username)
		return nil,&utils.NotFoundError{errMsg}
	}

	if (user.Credits - credits) < 0 {
		return nil, &utils.ForbiddenError{"User don't have enough points"}
	}

	user.Credits = user.Credits - credits
	tx.NamedExec("UPDATE users SET credits = :credits WHERE username = :username", &user)
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	return &user, nil
}
