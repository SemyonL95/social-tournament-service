package models

type Backer struct {
	ID           int    `db:"id"`
	PlayerID     string `db:"player_id"`
	BackerID     string `db:"backer_id"`
	TournamentID int    `db:"tournament_id"`
}
