package models

type Baker struct {
	ID           int     `database:"id"`
	PlayerId     int     `database:"player_id"`
	BakerId      int     `database:"baker_id"`
	TournamentID float64 `database:"tournament_id"`
}
