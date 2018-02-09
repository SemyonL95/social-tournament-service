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

func (db *DB) FundOrCreateUser(playerID string, credits float64) error {
	sql := `INSERT INTO users (username, credits) VALUES ($1, $2) 
			ON CONFLICT (username) DO UPDATE SET credits = $2;`

	_, err := db.conn.Exec(sql, playerID, credits)

	if err != nil {
		log.Println(err.Error())
		return err
	}

	return nil
}

func (db *DB) TakePointsFromUser(playerID string, credits float64) error {
	tx, err := db.conn.Beginx()
	if err != nil {
		log.Println(err)
		return err
	}
	user := models.User{}

	err = tx.Get(&user, `SELECT * FROM users WHERE username = $1 FOR UPDATE`, playerID)
	if err != nil {
		return err
	}

	if (user.Credits - credits) < 0 {
		return ErrNotEnoughMoney
	}

	user.Credits = user.Credits - credits
	tx.NamedExec(`UPDATE users SET credits = :credits WHERE username = :playerID`, &user)
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

func (db *DB) AssignToTournament(tournamentID int, playerID string, bakersIDs []string) error {
	tx, err := db.conn.Beginx()
	if err != nil {
		return err
	}

	tournament := models.Tournament{}
	err = tx.Get(&tournament, "SELECT * FROM tournaments WHERE id = $1", tournamentID)
	if err != nil {
		return err
	}

	var participants []models.User
	if bakersIDs != nil {
		queryStr := ""
		var tempBakersIDs []interface{}

		for i, _ := range bakersIDs {
			queryStr += fmt.Sprint("$", i + 1, ",")
			tempBakersIDs = append(tempBakersIDs, bakersIDs[i])
		}
		queryStr = queryStr[0:len(queryStr) - 1]
		sql := fmt.Sprintf("SELECT * FROM users WHERE username IN (%s) FOR UPDATE", queryStr)
		err := tx.Select(&participants, sql, tempBakersIDs...)
		if err != nil {
			log.Println("0")
			return err
		}
	}

	player := models.User{}
	err = tx.Get(&player, "SELECT * FROM users WHERE username = $1 FOR UPDATE", playerID)
	if err != nil {
		log.Println("1")
		return err
	}

	participants = append(participants, player)
	tournamentDeposit := tournament.Deposit / float64(len(participants))
	log.Printf("tournament deposit from 1 user %d", tournamentDeposit)
	stmtUsers, err := tx.Prepare("UPDATE users SET credits = (credits - $1) WHERE username = $2")
	if err != nil {
		log.Println("2")
		return err
	}
	stmtBakers, err := tx.Prepare("INSERT INTO bakers (player_id, baker_id, tournamen_id) VALUES ($1, $2, $3)")
	if err != nil {
		log.Println("3")
		return err
	}

	for _, participant := range participants {
		if participant.Credits < tournamentDeposit {
			return ErrNotEnoughMoney
		}

		if participant.Username == playerID {
			_, err = stmtUsers.Exec(tournamentDeposit, participant.Username)
			if err != nil {
				log.Println("4")
				return err
			}
			_, err := tx.Exec("INSERT INTO players (user_id, tournamen_id) VALUES ($1, $2)", tournament.ID, player.ID)
			if err != nil {
				log.Println("5")
				return err
			}
		} else {
			_, err = stmtUsers.Exec(tournamentDeposit, participant.Username)
			if err != nil {
				log.Println("6")
				return err
			}
			_, err := stmtBakers.Exec(player.ID, participant.ID, tournament.ID)
			if err != nil {
				log.Println("7")
				return err
			}
		}
	}


	err = tx.Commit()
	if err != nil {
		log.Println("8")
		return err
	}


	return nil
}
