package database

import (
	"fmt"
	"log"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/SemyonL95/social-tournament-service/src/models"
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

func (db *DB) FundOrCreateUser(playerID string, points float64) error {
	sqlInsertUsers := `INSERT INTO users (id, points) VALUES ($1, $2) 
			ON CONFLICT (id) DO UPDATE SET points = $2;`

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

func (db *DB) AssignToTournament(tournamentID int, playerID string, backersIDs []string) error {
	return retryTransact(db.conn, db.txAssignToTournament(tournamentID, playerID, backersIDs), 5)
}

func (db *DB) FinishTournament(winners models.Winners) error {
	return retryTransact(db.conn, db.txFinishTournament(winners), 5)
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

func (db *DB) txAssignToTournament(tournamentID int, playerID string, backersIDs []string) func(*sqlx.Tx) error {
	return func(tx *sqlx.Tx) error {
		var err error
		var stmtBackers *sql.Stmt

		tournament := models.Tournament{}
		err = tx.Get(&tournament, "SELECT * FROM tournaments WHERE id = $1 AND finished = FALSE", tournamentID)
		if err != nil {
			return err
		}

		existsPlayer := models.Player{}
		err = tx.Get(&existsPlayer, "SELECT * FROM players WHERE user_id = $1", playerID)
		if existsPlayer.ID != 0 {
			return ErrUserAlreadyJoined
		}

		var participants []models.User
		if backersIDs != nil {
			sqlSelectUsers, tempBackersIDs := queryBuildWhereInSelectUsers(backersIDs)
			err := tx.Select(&participants, sqlSelectUsers, tempBackersIDs...)
			if err != nil {
				return err
			}

			stmtBackers, err = tx.Prepare("INSERT INTO backers (player_id, backer_id, tournament_id) VALUES ($1, $2, $3)")
			if err != nil {
				return err
			}
			defer stmtBackers.Close()
		}

		player := models.User{}
		err = tx.Get(&player, querySelectUserForUpdate, playerID)
		if err != nil {
			return err
		}

		participants = append(participants, player)
		tournamentDeposit := tournament.Deposit / float64(len(participants))
		stmtUsers, err := tx.PrepareNamed(queryUpdateUsersCredits)
		if err != nil {
			return err
		}
		defer stmtUsers.Close()

		if err != nil {
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
				return err
			}

			if participant.ID == playerID {
				_, err := tx.Exec("INSERT INTO players (user_id, tournament_id) VALUES ($1, $2)", player.ID, tournament.ID)
				if err != nil {
					return err
				}
			} else {
				_, err := stmtBackers.Exec(player.ID, participant.ID, tournament.ID)
				if err != nil {
					return err
				}
			}
		}

		return nil
	}

}

func (db *DB) txFinishTournament(winners models.Winners) func(*sqlx.Tx) error {
	return func(tx *sqlx.Tx) error {
		tournament := models.Tournament{}
		err := tx.Get(&tournament, "SELECT * FROM tournaments WHERE id = $1 AND finished = FALSE", winners.TournamentID)
		if err != nil {
			return err
		}

		querySelectUsers := `SELECT users.id, users.points FROM players
 				  LEFT JOIN users ON players.user_id = users.id 
 				  WHERE players.tournament_id = $1 AND players.user_id = $2
				  	UNION ALL
					  SELECT users.id, users.points FROM backers
					  LEFT JOIN users ON backers.backer_id = users.id
					  WHERE backers.tournament_id = $1 AND backers.player_id = $2`

		stmtSelectUsers, err := tx.Preparex(querySelectUsers)
		if err != nil {
			return nil
		}
		defer stmtSelectUsers.Close()
		stmtUpdateUsers, err := tx.PrepareNamed(queryUpdateUsersCredits)
		if err != nil {
			return err
		}
		defer stmtUpdateUsers.Close()
		stmtCreateResult, err := tx.Preparex(`INSERT INTO results (winner_id, tournament_id, prize) VALUES ($1, $2, $3)`)
		if err != nil {
			return err
		}
		defer stmtCreateResult.Close()

		for _, winner := range winners.Winners {
			var users []models.User
			err = stmtSelectUsers.Select(&users, winners.TournamentID, winner.PlayerID)
			if err != nil {
				log.Println(err)
				return err
			}

			prize := winner.Prize / float64(len(users))
			for _, user := range users {
				user.Points += prize
				_, err := stmtUpdateUsers.Exec(&user)
				if err != nil {
					return err
				}
			}
			_, err = stmtCreateResult.Exec(winner.PlayerID, winners.TournamentID, winner.Prize, )
			if err != nil {
				return err
			}
		}

		_, err = tx.Exec("UPDATE tournaments SET finished = TRUE WHERE id = $1", winners.TournamentID)
		if err != nil {
			return err
		}

		return nil
	}
}

func queryBuildWhereInSelectUsers(IDs []string) (string, []interface{}) {
	queryStr := ""
	var tempBackersIDs []interface{}

	for i, _ := range IDs {
		queryStr += fmt.Sprint("$", i+1, ",")
		tempBackersIDs = append(tempBackersIDs, IDs[i])
	}
	queryStr = queryStr[0: len(queryStr)-1]
	sqlSelectUsers := fmt.Sprintf("SELECT * FROM users WHERE id IN (%s) FOR UPDATE", queryStr)

	return sqlSelectUsers, tempBackersIDs
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
		if !ok {
			return err
		}

		//40001 serialization_failure
		//40P01 deadlock_detected
		if driverError.Code != "40001" && driverError.Code != "40P01" {
			return err
		}
	}
	return err
}
