package models

type Tournament struct {
	Id int
	Deposit int
	WinnerId int `db: "winner_id"`
}