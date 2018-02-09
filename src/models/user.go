package models

type User struct {
	ID       int     `database:"id"`
	Username string  `database:"username"`
	Credits  float64 `database:"credits"`
}