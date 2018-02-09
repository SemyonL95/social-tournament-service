package models

type Tournament struct {
	ID      int     `db:"id"`
	Deposit float64 `db:"deposit"`
}