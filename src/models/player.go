package models

type Player struct {
	ID           int `database:"id"`
	UserID       int `database:"user_id"`
	TournamentID int `database:"tournament_id"`
}
