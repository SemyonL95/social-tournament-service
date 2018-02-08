package models

type User struct {
	Id int `database:"id"`
	Username string `database:"username"`
	Credits float64 `database:"credits"`
}