package models

type Tournament struct {
	ID       int     `db:"id"`
	Finished bool    `db:"finished"`
	Deposit  float64 `db:"deposit"`
}
