package models

type Tournament struct {
	Model
	Id int
	Deposit int
	WinnerId int `db: "winner_id"`
}