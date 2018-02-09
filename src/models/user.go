package models

type User struct {
	ID     string  `db:"id"`
	Points float64 `db:"points"`
}
