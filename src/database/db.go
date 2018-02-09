package database

import (
	"fmt"
	"log"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/SemyonL95/social-tournament-service/src/models"
	"database/sql"
	"github.com/lib/pq"
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

func (db *DB) FundOrCreateUser(playerID string, points float64) error {
	sqlInsertUsers := `INSERT INTO users (username, points) VALUES ($1, $2) 
			ON CONFLICT (username) DO UPDATE SET points = $2;`

	_, err := db.conn.Exec(sqlInsertUsers, playerID, points)

	if err != nil {
		log.Println(err.Error())
		return err
	}

	return nil
}

func (db *DB) TakePointsFromUser(playerID string, points float64) error {
	return retryTransact(db.conn, db.txTakePointsFromUser(playerID, points), 5)
}

func (db *DB) CreateTournament(id int, deposit float64) error {
	sqlInsertTournaments := `INSERT INTO tournaments (id, deposit) VALUES ($1, $2)`

	_, err := db.conn.Exec(sqlInsertTournaments, id, deposit)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) AssignToTournament(tournamentID int, playerID string, bakersIDs []string) error {
	return retryTransact(db.conn, db.txAssignToTournament(tournamentID, playerID, bakersIDs), 5)
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
		tx.NamedExec(queryUpdateUsersCredits, &user)

		return nil
	}
}

func (db *DB) txAssignToTournament(tournamentID int, playerID string, bakersIDs []string) func(*sqlx.Tx) error {
	return func(tx *sqlx.Tx) error {
		var err error
		var stmtBakers *sql.Stmt
		tournament := models.Tournament{}
		err = tx.Get(&tournament, "SELECT * FROM tournaments WHERE id = $1", tournamentID)
		if err != nil {
			return err
		}

		var participants []models.User
		if bakersIDs != nil {
			sqlSelectUsers, tempBakersIDs := queryBuildWhereInSelectUsers(bakersIDs)
			err := tx.Select(&participants, sqlSelectUsers, tempBakersIDs...)
			if err != nil {
				log.Println("0")
				return err
			}

			stmtBakers, err = tx.Prepare("INSERT INTO bakers (player_id, baker_id, tournamen_id) VALUES ($1, $2, $3)")
			if err != nil {
				return err
			}
			defer stmtBakers.Close()
		}

		player := models.User{}
		err = tx.Get(&player, querySelectUserForUpdate, playerID)
		if err != nil {
			log.Println("1")
			return err
		}

		participants = append(participants, player)
		tournamentDeposit := tournament.Deposit / float64(len(participants))
		log.Println(tournamentDeposit)
		stmtUsers, err := tx.PrepareNamed(queryUpdateUsersCredits)
		if err != nil {
			log.Println("2")
			return err
		}
		defer stmtUsers.Close()

		if err != nil {
			log.Println("3")
			return err
		}

		for _, participant := range participants {
			if participant.Points < tournamentDeposit {
				log.Printf("%f %f", participant.Points, tournamentDeposit)
				return ErrNotEnoughMoney
			}

			participant.Points -= tournamentDeposit
			_, err = stmtUsers.Exec(&participant)
			if err != nil {
				log.Println("4")
				return err
			}

			if participant.Username == playerID {
				_, err := tx.Exec("INSERT INTO players (user_id, tournamen_id) VALUES ($1, $2)", player.ID, tournament.ID)
				if err != nil {
					log.Println("5")
					return err
				}
			} else {
				_, err := stmtBakers.Exec(player.ID, participant.ID, tournament.ID)
				if err != nil {
					log.Println("7")
					return err
				}
			}
		}

		return nil
	}

}

func queryBuildWhereInSelectUsers(IDs []string) (string, []interface{}) {
	queryStr := ""
	var tempBakersIDs []interface{}

	for i, _ := range IDs {
		queryStr += fmt.Sprint("$", i+1, ",")
		tempBakersIDs = append(tempBakersIDs, IDs[i])
	}
	queryStr = queryStr[0:len(queryStr)-1]
	sqlSelectUsers := fmt.Sprintf("SELECT * FROM users WHERE username IN (%s) FOR UPDATE", queryStr)

	return sqlSelectUsers, tempBakersIDs
}

func transact(db *sqlx.DB, txFunc func(*sqlx.Tx) error) (err error) {
	tx, err := db.Beginx()
	if err != nil {
		return
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // re-throw panic after Rollback
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
			if err != nil {
				err = tx.Rollback()
			}
		}
	}()
	err = txFunc(tx)
	return err
}

//We retry on those two error codes. This is due to serializable isolation level transactions set
func retryTransact(db *sqlx.DB, txFunc func(*sqlx.Tx) error, retryNumber int) error {
	var err error
	for i := 0; i < retryNumber; i++ {
		err = transact(db, txFunc)
		if err == nil {
			return err
		}
		driverError, ok := err.(*pq.Error)
		if !ok{
			return err
		}

		//40001 serialization_failure
		//40P01 deadlock_detected
		if  driverError.Code != "40001" && driverError.Code != "40P01"{
			return err
		}
	}
	return err
}