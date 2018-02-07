package models

type Player struct {
	Model
	Id int
	UserId int `db:user_id`
	Deposit float64
}
