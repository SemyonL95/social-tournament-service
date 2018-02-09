package models

type Player struct {
	ID           int    `db:"id"`
	UserID       string `db:"user_id"`
	TournamentID int    `db:"tournament_id"`
}
