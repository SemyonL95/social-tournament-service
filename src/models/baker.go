package models

type Baker struct {
	Id int
	PlayerId int `database: "player_id"`
	BakerId int `database: "baker_id"`
	Deposit float64
}