package models

type Result struct {
	Id int
	WinnerId int `db: "winner_id"`
	Prize float64
}