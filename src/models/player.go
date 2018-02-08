package models

type Player struct {
	Id int
	UserId int `database: "user_id"`
	Deposit float64
}
