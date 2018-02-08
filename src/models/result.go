package models

type Result struct {
	Id int
	WinnerId int `database: "winner_id"`
	Prize float64
}