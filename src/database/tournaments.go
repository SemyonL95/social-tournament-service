package database

import (
	"log"
	"database/sql"

	"github.com/jmoiron/sqlx"

	"github.com/SemyonL95/social-tournament-service/src/models"
)

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
				return err
			}

			usersCount := len(users)
			if usersCount < 1 {
				return ErrNoPlayersInTournament
			}

			prize := winner.Prize / float64(usersCount)
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
