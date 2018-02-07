package models

type Baker struct {
	Model
	Id int
	PlayerId int `db:player_id`
	BakerId int `db:baker_id`
	Deposit float64
}