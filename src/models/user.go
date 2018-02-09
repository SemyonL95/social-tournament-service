package models

type User struct {
	ID       int     `database:"id"`
	Username string  `database:"username"`
	Points   float64 `database:"points"`
}