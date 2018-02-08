package models

type User struct {
	Id int `db:"id"`
	Username string `db:"username"`
	Credits float64 `db:"credits"`
}