package models

type Winners struct {
	TournamentID int      `json:"tournamentId"`
	Winners      []Winner `json:"winners"`
}

type Winner struct {
	PlayerID string  `json:"playerId"`
	Prize    float64 `json:"prize"`
}
