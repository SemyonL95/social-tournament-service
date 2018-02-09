package models

type Result struct {
	ID       int
	WinnerId int `database: "winner_id"`
	Prize    float64
}