package models

type Player struct {
	Id int
	UserId int `db: "user_id"`
	Deposit float64
}
