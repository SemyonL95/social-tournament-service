package models

type Result struct {
	ID           int     `db:"id"`
	WinnerID     string  `db:"winner_id"`
	TournamentID int     `db:"tournament_id"`
	Prize        float64 `db:"prize"`
}
