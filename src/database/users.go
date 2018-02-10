package database

import (
	"github.com/jmoiron/sqlx"

	"github.com/SemyonL95/social-tournament-service/src/models"
)

func (db *DB) FundOrCreateUser(playerID string, points float64) error {
	sqlInsertUsers := `INSERT INTO users (id, points) VALUES ($1, $2) 
			ON CONFLICT (id) DO UPDATE SET points = $2;`

	_, err := db.conn.Exec(sqlInsertUsers, playerID, points)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) TakePointsFromUser(playerID string, points float64) error {
	return retryTransact(db.conn, db.txTakePointsFromUser(playerID, points), 5)
}

func (db *DB) GetUserBalance(playerID string) (*models.User, error) {
	user := models.User{}
	err := db.conn.Get(&user, "SELECT * FROM users WHERE id = $1", playerID)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (db *DB) txTakePointsFromUser(playerID string, points float64) func(*sqlx.Tx) error {
	return func(tx *sqlx.Tx) error {
		user := models.User{}

		err := tx.Get(&user, querySelectUserForUpdate, playerID)
		if err != nil {
			return err
		}

		if (user.Points - points) < 0 {
			return ErrNotEnoughMoney
		}

		user.Points -= points
		_, err = tx.NamedExec(queryUpdateUsersCredits, &user)
		if err != nil {
			return err
		}
		return nil
	}
}
